package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"net/url"
	"os"
)

type CreateResponse struct {
	Shortcode   string  `json:"shortcode"`
	Shortlink   string  `json:"shortlink"`
	Destination string  `json:"destination"`
	Secret      string  `json:"secret"`
	ExpiresAt   *string `json:"expiresAt"`
}

type ErrorResponse struct {
	RequestID string `json:"requestId"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
}

const CreateEndpoint = "https://go.super.fish/api/v1/shortlinks"

func main() {
	var apiKey string

	var rootCmd = &cobra.Command{
		Use:   "tetra [long-url]",
		Short: "Creates a shortlink from a long URL",
		Long:  "Creates a shortlink from a long URL",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				cmd.Help()
				os.Exit(0)
			}

			longUrl := args[0]
			if !isValidURL(longUrl) {
				println("Invalid URL")
				os.Exit(1)
			}

			shortlink, err := makeCreateRequest(longUrl, apiKey)
			if err != nil {
				fmt.Printf("Error creating shortlink: %s\n", err)
				os.Exit(1)
			}

			println("\n" + shortlink + "\n")
		},
	}

	rootCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "API key")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func isValidURL(input string) bool {
	u, err := url.Parse(input)
	if err != nil {
		return false
	}

	return u.Scheme == "https" || u.Scheme == "http"
}

func makeCreateRequest(longUrl string, apiKey string) (string, error) {
	reqBody := bytes.NewReader([]byte(fmt.Sprintf(`{"destination": "%s"}`, longUrl)))

	reqUrl := CreateEndpoint
	if apiKey != "" {
		reqUrl = fmt.Sprintf("%s?apiKey=%s", reqUrl, apiKey)
	}

	req, err := http.NewRequest(http.MethodPut, reqUrl, reqBody)
	if err != nil {
		return "", err
	}

	req.Header.Add("User-Agent", "Tetra CLI")
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 201 {
		var errBody ErrorResponse
		if err := json.Unmarshal(body, &errBody); err != nil {
			return "", fmt.Errorf("failed to parse error response: %s", err)
		}

		return "", fmt.Errorf("%s (%d)", errBody.Message, errBody.Code)
	}

	var resBody CreateResponse
	if err := json.Unmarshal(body, &resBody); err != nil {
		return "", err
	}

	return resBody.Shortlink, nil
}
