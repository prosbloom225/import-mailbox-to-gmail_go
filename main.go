package main

import (
	"bytes"
	"fmt"
	"github.com/bthomson/mbox"
	"github.com/kalaspuffar/base64url"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/gmail/v1"
	"io/ioutil"
	"net/mail"
	"os"
	"time"
	"unsafe"
)

var SVC_ACCT_EMAIL = "524816920325-tj75v0nolh6dr22fafodpbncma25dgkn.apps.googleusercontent.com"
var SVC_ACCT_PRIVATE_KEY_PATH = "/mnt/mbox/import-mailbox-to-gmail_go/private_key.pem"
var LABEL_NAME = "import_label"
var MBOX_DIR = "./mbox/"

var WORKER_MBOX_THREADS = 3

func Build_Gmail_Service(sub_user string) (*gmail.Service, error) {
	buf, err := ioutil.ReadFile(SVC_ACCT_PRIVATE_KEY_PATH)
	if err != nil {
		fmt.Printf("Error loading private key: %v", err)
		return nil, err
	} else {
		conf := &jwt.Config{
			Email:      SVC_ACCT_EMAIL,
			PrivateKey: buf,
			Scopes: []string{
				"https://www.googleapis.com/auth/gmail.modify",
			},
			TokenURL: google.JWTTokenURL,
			Subject:  sub_user,
		}
		client := conf.Client(oauth2.NoContext)
		client.Get("...")

		svc, err := gmail.New(client)
		if err != nil {
			fmt.Printf("Unable to retrieve gmail client: %v\n", err)
			return nil, err
		}
		fmt.Printf("Gmail service initialized for user: %s\n", sub_user)
		return svc, nil
	}
}

func Create_Label(svc *gmail.Service, labelName string) (*string, error) {
	label := gmail.Label{
		Name: labelName,
		// TODO - this is off to show the label in the inbox list for easier deletion
		// LabelListVisibility: "labelHide",
		MessageListVisibility: "hide",
	}
	ret, err := svc.Users.Labels.Create("me", &label).Do()
	if err != nil {
		//TODO - if label exists, get that one and return it
		return nil, err
	} else {
		return &ret.Id, nil
	}
}

func Process_Mbox(filename string) ([]*mail.Message, error) {
	fmt.Printf("Processing mbox: %s\n", filename)
	r, err := mbox.ReadFile(filename, false)
	if err != nil {
		fmt.Printf("Error reading mbox file: %v\n", err)
		return nil, err
	} else {
		return r, nil
	}
}

func Worker_Mbox(id int, labelId string, mboxes <-chan os.FileInfo, results chan<- int) {
	for j := range mboxes {
		fmt.Println("worker", id, "label", labelId, "processing mbox", j.Name())
		time.Sleep(time.Second)
		results <- 1
		/*
			msgs, err := Process_Mbox(MBOX_DIR + userName + "/" + mbox.Name())
			if err == nil {
			  // Loop through each msg in the mbox and import
			  for _, msg := range msgs {
				var reformed_msg string
				for k, v := range msg.Header {
				  reformed_msg += fmt.Sprintf("%s: %s\r\n", k, v[0])
				}
				buf := new(bytes.Buffer)
				buf.ReadFrom(msg.Body)
				b := buf.Bytes()
				s := *(*string)(unsafe.Pointer(&b))
				reformed_msg += "\r\n" + s

				// TODO - retry logic/incremental backoff
				gmsg := gmail.Message{
				  Raw:      base64url.Encode([]byte(reformed_msg)),
				  LabelIds: []string{*labelId},
				}
				_, err = svc.Users.Messages.Insert("me", &gmsg).Do()
				if err != nil {
				  fmt.Printf("ERROR: %v", err)
				} else {
				  fmt.Println("Email imported successfully")
				}
				// TODO - Only run a single email
				// os.Exit(1)
			  }
		*/
	}
}

func Worker_User(userName string) {
	fmt.Printf("Processing user: %s\n", userName)
	svc, err := Build_Gmail_Service(userName)
	if err != nil {
		fmt.Printf("Error building gmail service: %v", err)
	} else {
		// Generate a label
		labelId, err := Create_Label(svc, LABEL_NAME)
		if err != nil {
			fmt.Printf("Error creating label: %v", err)
		} else {
			// Get list of mboxes
			files, err := ioutil.ReadDir(MBOX_DIR + userName)
			if err != nil {
				fmt.Printf("Could not parse directory: ", err)
			} else {
				// Build queues
				mboxes := make(chan os.FileInfo, 100)
				results := make(chan int, 100)
				// Spin up worker_mbox threads
				for w := 1; w <= WORKER_MBOX_THREADS; w++ {
					go Worker_Mbox(w, *labelId, mboxes, results)
				}

				// Iterate mboxes
				for _, mbox := range files {
					fmt.Printf("Mbox: %s\n", mbox.Name())
					mboxes <- mbox
				}
				close(mboxes)
				fmt.Println("Waiting...")

				for a := 1; a <= 9; a++ {
					<-results
				}
				os.Exit(1)

			}
		}
	}
}

func main() {
	fmt.Print("BEGIN\n")

	user_id := "angela.starke@kohls.com"
	Worker_User(user_id)
	os.Exit(1)

	svc, err := Build_Gmail_Service(user_id)
	labelId, err := Create_Label(svc, "testLabel")
	if err != nil {
		fmt.Printf("ERROR creating label: %v", err)
	} else {
		fmt.Printf("LabelId: %s", *labelId)
	}
	os.Exit(1)

	// get label id for test
	// ret, err := svc.Users.Labels.Get("Name", "test").Do()
	/*
		  // List
		  ret, err := svc.Users.Labels.List("angela.starke@kohls.com").Do()
		  if err != nil {
			fmt.Printf("ERROR GETTING LABEL: %v", err)
		  } else {
			for _, label := range ret.Labels {
			  fmt.Printf("Label: %s,%s\n", label.Name, label.Id)
			}
		  }
	*/
	ret, err := svc.Users.Labels.Get("me", "Label_110").Do()
	if err != nil {
		fmt.Printf("ERROR GETTING LABEL: %v", err)
	} else {
		fmt.Printf("Label: %s", ret.Name)
		fmt.Printf("Id: %s", ret.Id)
	}
	os.Exit(1)

	msgs, err := Process_Mbox("/mnt/mbox/import-mailbox-to-gmail_go/mbox/angela.starke@kohls.com/Export_angela.mayweatherskohls.com_202458127_2012_1.mbox")
	if err == nil {
		for _, msg := range msgs {
			var reformed_msg string
			reformed_msg = fmt.Sprintf("%s: %s\r\n", "labelIds", "test")
			for k, v := range msg.Header {
				reformed_msg += fmt.Sprintf("%s: %s\r\n", k, v[0])
			}
			// fmt.Printf("%s\n", reformed_msg)
			// os.Exit(1)

			buf := new(bytes.Buffer)
			buf.ReadFrom(msg.Body)
			b := buf.Bytes()
			s := *(*string)(unsafe.Pointer(&b))
			reformed_msg += "\r\n" + s

			gmsg := gmail.Message{
				Raw:      base64url.Encode([]byte(reformed_msg)),
				LabelIds: []string{"test"},
			}
			fmt.Printf("%s\n", reformed_msg)
			_, err = svc.Users.Messages.Insert("me", &gmsg).Do()
			if err != nil {
				fmt.Printf("ERROR: %v", err)
			}
			// TODO - Only run a single email
			os.Exit(1)
		}
	}

	// TEST GMAIL
	// user_id := "michael.osiecki@kohls.com"
	// svc, err := Build_Gmail_Service(user_id)
	// if err != nil {
	// 	fmt.Printf("Could not initialize gmail service: %v\n", err)
	// } else {
	// 	r, err := svc.Users.Labels.List(user_id).Do()
	// 	if err != nil {
	// 		fmt.Printf("Error getting labels: %v", err)
	// 	} else {
	// 		if len(r.Labels) > 0 {
	// 			fmt.Print("Labels:\n")
	// 			for _, l := range r.Labels {
	// 				fmt.Printf("- %s\n", l.Name)
	// 			}
	// 		} else {
	// 			fmt.Printf("No labels found\n")
	// 		}
	// 	}
	// }
}
