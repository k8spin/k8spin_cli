package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/parnurzeal/gorequest"
	"github.com/urfave/cli"
)

type namespace struct {
	Namespace        string   `json:"namespace"`
	NamespaceName    string   `json:"namespace_name"`
	IngressWhitelist []string `json:"ingress_whitelist"`
	Expiration       string   `json:"expiration"`
	ResourceQuota    string   `json:"resource_quotas"`
	Status           string   `json:"status"`
}

var Api_Base string = "https://console.beta.k8spin.cloud/api"
var debug bool
var token string

func main() {

	app := cli.NewApp()
	app.Name = "k8spin"
	app.Version = "0.1.0"
	app.Usage = "CLI for managing namespaces"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "show debug information",
			Destination: &debug,
		},
		cli.StringFlag{
			Name:        "token",
			Usage:       "K8Spin token",
			EnvVar:      "K8SPINTOKEN",
			Destination: &token,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list all namespaces",
			Action: func(c *cli.Context) error {
				request := gorequest.New()
				resp, body, _ := request.Get(Api_Base+"/namespaces").Set("K8SPIN-TOKEN", token).End()
				debugRequest(body)
				if httpCodeCheck(resp.StatusCode) {
					printNamespacesTable(body)
				}
				return nil
			},
		},
		{
			Name:    "get_credentials",
			Aliases: []string{"gc"},
			Usage:   "get namespace credentials",
			Action: func(c *cli.Context) error {
				var namespace = c.Args().First()

				request := gorequest.New()
				resp, body, _ := request.Get(Api_Base+"/namespaces/"+namespace).Set("K8SPIN-TOKEN", token).End()
				debugRequest(body)
				if httpCodeCheck(resp.StatusCode) {
					fmt.Println(body)
				}
				return nil
			},
		},
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "create a namespace",
			Action: func(c *cli.Context) error {
				var namespace = c.Args().First()

				request := gorequest.New()
				resp, body, _ := request.Post(Api_Base+"/namespaces").
					Set("K8SPIN-TOKEN", token).
					Send(`{"namespace_name":"` + namespace + `"}`).
					End()

				debugRequest(body)
				if httpCodeCheck(resp.StatusCode) {
					fmt.Println("Namespace " + namespace + " created")
				}
				return nil
			},
		},
		{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "delete a namespace",
			Action: func(c *cli.Context) error {
				var namespace = c.Args().First()

				request := gorequest.New()
				resp, body, _ := request.Delete(Api_Base+"/namespaces/"+namespace).
					Set("K8SPIN-TOKEN", token).
					End()

				debugRequest(body)
				if httpCodeCheck(resp.StatusCode) {
					fmt.Println("Namespace " + namespace + " deleted")
				}
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func debugRequest(body string) {
	if debug {
		fmt.Println("Response body :")
		fmt.Println(body)
	}
}

func httpCodeCheck(httpCode int) bool {
	if debug {
		fmt.Println("HTTP Code:")
		fmt.Println(httpCode)
	}
	if httpCode == 401 {
		fmt.Println("API token not valid")
		return false
	} else if httpCode == 404 {
		fmt.Println("Not found")
		return false
	}
	return true
}

func printNamespacesTable(body string) {
	namespaces := []namespace{}
	json.Unmarshal([]byte(body), &namespaces)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Namespace", "Expiration", "Status"})
	for _, namespace := range namespaces {
		table.Append([]string{namespace.NamespaceName, namespace.Namespace, namespace.Expiration, namespace.Status})
	}
	table.Render() // Send output
}
