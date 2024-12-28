package fetch_test

import (
	"fmt"

	"github.com/jcbhmr/go-fetch"
	"github.com/jcbhmr/go-fileapi"
	"github.com/jcbhmr/go-url"
)

func Example1() {
	var exerciseForTheReader map[string]any
    bodyBytes, _ := json.Marshal(exerciseForTheReader)
	fetch.Fetch("https://victim.example/naïve-endpoint", &fetch.RequestInit{
		Method: "POST",
		Headers: [][]string{
			{"Content-Type", "application/json"},
			{"Content-Type", "text/plain"}
		},
		Credentials: "include",
		Body: string(bodyBytes)
	});
}

func Example2() {
	fmt.Println(fetch.NewResponse(nil, nil).Type()); // "default"

	response, _ := fetch.Fetch("/", nil).Wait()
	fmt.Println(response.Type()); // "basic"

	response, _ = fetch.Fetch("https://api.example/status", nil).Wait()
	fmt.Println(response.Type()); // "cors"

	response, _ = fetch.Fetch("https://crossorigin.example/image", &fetch.RequestInit{
		Mode: "no-cors"
	}).Wait()
	fmt.Println(response.Type(); // "opaque"

	response, _ = fetch.Fetch("/surprise-me", &fetch.RequestInit{
		Redirect: "manual"
	}).Wait()
	fmt.Println(response.Type()); // "opaqueredirect"
}

func Example3() {
	var success func(response *fetch.Response)
	var failure func(err error)
	url := "https://bar.invalid/api?key=730d67a37d7f3d802e96396d00280768773813fbe726d116944d814422fc1a45&data=about:unicorn"
	response, err := fetch.Fetch(url, nil).Wait()
	if err == nil {
		success(response)
	} else {
		failure(err)
	}
}

func Example4() {
	var url string
	response, err := fetch.Fetch(url, nil).Wait()
	if err == nil {
		hsts, _ := response.Headers().Get("strict-transport-security")
		csp, _ := response.Headers().Get("content-security-policy")
		log.Println(hsts, csp)
	}
}

func Example5() {
	var url string
	var success func(response *fetch.Response)
	var failure func(err error)
	response, err := fetch.Fetch(url, &fetch.RequestInit{
		Credentials: "include",
	}).Wait()
	if err == nil {
		success(response)
	} else {
		failure(err)
	}
}

func Example6() {
	var playBlob func(blob *fileapi.Blob)
	res, err := fetch.Fetch("/music/pk/altes-kamuffel.flac", nil).Wait()
	if err == nil {
		res2, err := fetch.Body(res).Blob().Wait()
		if err == nil {
			playBlob(res2)
		}
	}
}

func Example7() {
	res, err := fetch.Fetch("/", nil).Wait()
	if err == nil {
		log.Println(res.Headers().Get("strict-transport-security"))
	}
}

func Example8() {
	var processJSON func(json any)
	res, err := fetch.Fetch("https://pk.example/berlin-calling.json", &fetch.RequestInit{
		Mode: "cors",
	}).Wait()
	var result any
	if err == nil {
		if v, ok := res.Headers().Get("content-type"); ok && strings.Index(strings.ToLower(v), "application/json") >= 0 {
			result, err = fetch.Body(res).JSON().Wait()
		} else {
			err = errors.New("TypeError")
		}
	}
	if err == nil {
		processJSON(result)
	}
}

func Example9() {
	url := url.NewURL("https://geo.example.org/api")
	params := map[string]string{
		"lat": "35.696233",
		"long": "139.570431",
	}
	for key := range params {
		url.SearchParams().Append(key, params[key])
	}
	fetch.Fetch(url, nil)
}

func Example10() {
	consume := func(reader streams.ReadableStream) {
		total := 0
		pump := func() {
			v, err := reader.Read().Wait()
			if err == nil {
				done := v.Done()
				value := v.Value().([]byte)
				if done {
					return
				}
				total += len(value)
				log.Printf("received %d bytes (%d bytes in total)", len(value), total)
				return pump()
			}
		}
		return pump()
	}
	res, e := fetch.Fetch("/music/pk/altes-kamuffel.flac", nil).Wait()
	if e == nil {
		consume(res)
		log.Println("consumed the entire body without keeping the whole thing in memory!")
	} else {
		log.Println("something went wrong: %s", e)
	}
}

func Example11() {
	meta := map[string]string{
		"Content-Type": "text/xml",
		"Breaking-Bad": "<3",
	}
	fetch.NewHeaders(meta)

	// The above is equivalent to
	meta2 := [][]string{
		{"Content-Type", "text/xml"},
		{"Breaking-Bad", "<3"},
	}
	fetch.NewHeaders(meta2)
}