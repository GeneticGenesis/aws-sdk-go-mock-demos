package aws

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync"
	"testing"
)

func TestEC2Request(t *testing.T) {
	var m sync.Mutex
	var httpReq *http.Request
	var form url.Values

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			m.Lock()
			defer m.Unlock()

			httpReq = r

			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			form = r.Form

			fmt.Fprintln(w, `<Thing><IpAddress>woo</IpAddress></Thing>`)
		},
	))
	defer server.Close()

	client := EC2Client{
		Context: Context{
			Service: "animals",
			Region:  "us-west-2",
			Credentials: Creds(
				"accessKeyID",
				"secretAccessKey",
				"securityToken",
			),
		},
		Client:     http.DefaultClient,
		Endpoint:   server.URL,
		APIVersion: "1.1",
	}

	req := fakeEC2Request{
		PresentString:  String("string"),
		PresentBoolean: True(),
		PresentInteger: Integer(1),
		PresentLong:    Long(2),
		PresentDouble:  Double(1.2),
		PresentFloat:   Float(2.3),
		PresentSlice:   []string{"one", "two"},
		PresentStruct:  &EmbeddedStruct{Value: String("v")},
		PresentStructSlice: []EmbeddedStruct{
			{Value: String("p")},
			{Value: String("q")},
		},
	}
	var resp fakeEC2Response
	if err := client.Do("GetIP", "POST", "/", &req, &resp); err != nil {
		t.Fatal(err)
	}

	m.Lock()
	defer m.Unlock()

	if v, want := httpReq.Method, "POST"; v != want {
		t.Errorf("Method was %v but expected %v", v, want)
	}

	if httpReq.Header.Get("Authorization") == "" {
		t.Error("Authorization header is missing")
	}

	if v, want := httpReq.Header.Get("Content-Type"), "application/x-www-form-urlencoded"; v != want {
		t.Errorf("Content-Type was %v but expected %v", v, want)
	}

	if v, want := httpReq.Header.Get("User-Agent"), "aws-go"; v != want {
		t.Errorf("User-Agent was %v but expected %v", v, want)
	}

	if err := httpReq.ParseForm(); err != nil {
		t.Fatal(err)
	}

	expectedForm := url.Values{
		"Action":                     []string{"GetIP"},
		"Version":                    []string{"1.1"},
		"PresentString":              []string{"string"},
		"PresentBoolean":             []string{"true"},
		"PresentInteger":             []string{"1"},
		"PresentLong":                []string{"2"},
		"PresentDouble":              []string{"1.2"},
		"PresentFloat":               []string{"2.3"},
		"PresentSlice.1":             []string{"one"},
		"PresentSlice.2":             []string{"two"},
		"PresentStruct.Value":        []string{"v"},
		"PresentStructSlice.1.Value": []string{"p"},
		"PresentStructSlice.2.Value": []string{"q"},
	}

	if !reflect.DeepEqual(form, expectedForm) {
		t.Errorf("Post body was \n%s\n but expected \n%s", form.Encode(), expectedForm.Encode())
	}

	if want := (fakeEC2Response{IPAddress: "woo"}); want != resp {
		t.Errorf("Response was %#v, but expected %#v", resp, want)
	}
}

func TestEC2RequestError(t *testing.T) {
	var m sync.Mutex
	var httpReq *http.Request
	var form url.Values

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			m.Lock()
			defer m.Unlock()

			httpReq = r

			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			form = r.Form

			w.WriteHeader(400)
			fmt.Fprintln(w, `<Response>
<RequestId>woo</RequestId>
<Errors>
<Error>
<Type>Problem</Type>
<Code>Uh Oh</Code>
<Message>You done did it</Message>
</Error>
</Errors>
</Response>`)
		},
	))
	defer server.Close()

	client := EC2Client{
		Context: Context{
			Service: "animals",
			Region:  "us-west-2",
			Credentials: Creds(
				"accessKeyID",
				"secretAccessKey",
				"securityToken",
			),
		},
		Client:     http.DefaultClient,
		Endpoint:   server.URL,
		APIVersion: "1.1",
	}

	req := fakeEC2Request{}
	var resp fakeEC2Response
	err := client.Do("GetIP", "POST", "/", &req, &resp)
	if err == nil {
		t.Fatal("Expected an error but none was returned")
	}

	if err, ok := err.(APIError); ok {
		if v, want := err.Type, "Problem"; v != want {
			t.Errorf("Error type was %v, but expected %v", v, want)
		}

		if v, want := err.Code, "Uh Oh"; v != want {
			t.Errorf("Error type was %v, but expected %v", v, want)
		}

		if v, want := err.Message, "You done did it"; v != want {
			t.Errorf("Error message was %v, but expected %v", v, want)
		}
	} else {
		t.Errorf("Unknown error returned: %#v", err)
	}
}

type fakeEC2Request struct {
	PresentString StringValue `ec2:"PresentString"`
	MissingString StringValue `ec2:"MissingString"`

	PresentInteger IntegerValue `ec2:"PresentInteger"`
	MissingInteger IntegerValue `ec2:"MissingInteger"`

	PresentLong LongValue `ec2:"PresentLong"`
	MissingLong LongValue `ec2:"MissingLong"`

	PresentDouble DoubleValue `ec2:"PresentDouble"`
	MissingDouble DoubleValue `ec2:"MissingDouble"`

	PresentFloat FloatValue `ec2:"PresentFloat"`
	MissingFloat FloatValue `ec2:"MissingFloat"`

	PresentBoolean BooleanValue `ec2:"PresentBoolean"`
	MissingBoolean BooleanValue `ec2:"MissingBoolean"`

	PresentSlice []string `ec2:"PresentSlice"`
	MissingSlice []string `ec2:"MissingSlice"`

	PresentStructSlice []EmbeddedStruct `ec2:"PresentStructSlice"`
	MissingStructSlice []EmbeddedStruct `ec2:"MissingStructSlice"`

	PresentStruct *EmbeddedStruct `ec2:"PresentStruct"`
	MissingStruct *EmbeddedStruct `ec2:"MissingStruct"`
}

type fakeEC2Response struct {
	IPAddress string `xml:"IpAddress"`
}
