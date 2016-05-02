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
	"unsafe"
)

var SVC_ACCT_EMAIL = "524816920325-tj75v0nolh6dr22fafodpbncma25dgkn.apps.googleusercontent.com"
var SVC_ACCT_PRIVATE_KEY_PATH = "/mnt/mbox/import-mailbox-to-gmail_go/private_key.pem"

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

func Parse_Mbox(filename string) ([]*mail.Message, error) {
	r, err := mbox.ReadFile(filename, false)
	if err != nil {
		fmt.Printf("Error reading mbox file: %v", err)
		return nil, err
	} else {
		return r, nil
	}
}

func main() {
	fmt.Print("BEGIN\n")

	user_id := "angela.starke@kohls.com"
	svc, err := Build_Gmail_Service(user_id)

	msgs, err := Parse_Mbox("/mnt/mbox/import-mailbox-to-gmail_go/mbox/angela.starke@kohls.com/Export_angela.mayweatherskohls.com_202458127_2012_1.mbox")
	if err == nil {
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

			gmsg := gmail.Message{
				// Raw: encodeWeb64String([]byte(reformed_msg)),
				// Raw: b64.StdEncoding.EncodeToString([]byte(reformed_msg)),
				Raw: base64url.Encode([]byte(reformed_msg)),
				// Raw: reformed_msg,
			}
			fmt.Printf("%s\n", reformed_msg)
			_, err = svc.Users.Messages.Insert("me", &gmsg).Do()
			if err != nil {
				fmt.Printf("ERROR: %v", err)
			}

			// Only run a single email
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
