package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/netlify/open-api/go/plumbing"
	"github.com/netlify/open-api/go/plumbing/operations"
	"github.com/go-openapi/runtime"
	openapiClient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

)

// Netlify specific constants
const (
	NetlifyAPIHost string = "api.netlify.com"
	NetlifyAPIPath string = "/api/v1"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	fmt.Println("Finding deploy preview URL for commit:", request.QueryStringParameters["commit"])

	var deploys, error = getNetlifyClient().Operations.ListSiteDeploys(getListSiteDeploysParams(), getAuthInfo())
	fmt.Println("Deploys:", deploys.Payload)
	fmt.Println("Error:", error)

	const deploy_preview_url = "https://netlify-function--agilepathway-co-uk.netlify.com"
	fmt.Println("Deploy preview url found:", deploy_preview_url)
	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       deploy_preview_url,
	}, nil
}

func getListSiteDeploysParams() (*operations.ListSiteDeploysParams){
	// soon os.Getenv("SITE_ID") should be available - https://github.com/netlify/build/issues/743
	return operations.NewListSiteDeploysParams().WithSiteID(os.Getenv("AGILE_PATHWAY_SITE_ID")) 
}

func getNetlifyClient() (*plumbing.Netlify) {
	transport := openapiClient.NewWithClient(NetlifyAPIHost, NetlifyAPIPath, plumbing.DefaultSchemes, getHTTPClient())
	client := plumbing.New(transport, strfmt.Default)
	return client
}

func getAuthInfo() (runtime.ClientAuthInfoWriter){
	return runtime.ClientAuthInfoWriterFunc(func(r runtime.ClientRequest, _ strfmt.Registry) error {
		r.SetHeaderParam("User-Agent", "agilepathway")
		r.SetHeaderParam("Authorization", "Bearer "+os.Getenv("LIST_SITE_DEPLOYS_TOKEN"))
		return nil
	})
}

func getHTTPClient() *http.Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   -1,
			DisableKeepAlives:     true}}

	return httpClient
}

func main() {
	// Make the handler available for Remote Procedure Call by AWS Lambda
	lambda.Start(handler)
}
