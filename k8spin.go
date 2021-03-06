package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"

	"github.com/olekukonko/tablewriter"
	"github.com/parnurzeal/gorequest"
	"github.com/urfave/cli"
)

type httpError struct {
	Error string `json:"error"`
}

type namespace struct {
	Namespace        string   `json:"namespace"`
	NamespaceName    string   `json:"namespace_name"`
	IngressWhitelist []string `json:"ingress_whitelist"`
	Expiration       string   `json:"expiration"`
	ResourceQuota    string   `json:"resource_quotas"`
	Status           string   `json:"status"`
}

var api_base string = "https://console.beta.k8spin.cloud/api"
var debug bool
var token string
var set_config bool
var cpu_amount string
var memory_amount string
var storage_amount string

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	app := cli.NewApp()
	app.Name = "k8spin"
	app.Version = "1.1.0"
	app.Usage = "CLI for managing namespaces"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "show debug information",
			Destination: &debug,
		},
		cli.StringFlag{
			Name:        "host",
			Usage:       "K8Spin host",
			EnvVar:      "K8SPINHOST",
			Value:       "https://api.k8spin.cloud",
			Destination: &api_base,
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
				var url = api_base+"/namespaces"

				resp, body, _ := request.Get(url).
					Set("Authorization", "Bearer "+token).
					End()
				if httpCodeCheck(resp) {
					printNamespacesTable(body)
				}
				return nil
			},
		},
		{
			Name:      "get_credentials",
			Aliases:   []string{"gc"},
			Usage:     "print namespace credentials to STDOUT",
			ArgsUsage: "[name]",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "set_config",
					Usage:       "Saves and sets KUBECONFIG variable to a temporary file",
					Destination: &set_config,
				},
			},
			Action: func(c *cli.Context) error {
				var namespace = c.Args().First()
				if namespace == "" {
					fmt.Println("You need to select a namespace")
					return nil
				}
				var url = api_base+"/namespaces/"+namespace
				request := gorequest.New()

				resp, body, _ := request.Get(url).
					Set("Authorization", "Bearer "+token).
					End()
				if httpCodeCheck(resp) {
					fmt.Println(body)
				}
				return nil
			},
		},
		{
			Name:      "set_credentials",
			Aliases:   []string{"sc"},
			Usage:     "set credentials in a temp KUBECONFIG",
			ArgsUsage: "[name]",
			Action: func(c *cli.Context) error {
				var namespace = c.Args().First()
				if namespace == "" {
					fmt.Println("You need to select a namespace")
					return nil
				}
				var url = api_base+"/namespaces/"+namespace

				request := gorequest.New()
				resp, body, _ := request.Get(url).
					Set("Authorization", "Bearer "+token).
					End()
				if httpCodeCheck(resp) {
					var filePath = "/tmp/k8spin_" + namespace
					err := ioutil.WriteFile(filePath, []byte(body), 0644)
					check(err)
					os.Setenv("KUBECONFIG", filePath)
					fmt.Println("Credentials saved at " + filePath)
					fmt.Println("KUBECONFIG variable set")
					syscall.Exec(os.Getenv("SHELL"), []string{os.Getenv("SHELL")}, syscall.Environ())
				}
				return nil
			},
		},
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "create a namespace",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "cpu", Value: "100", Destination: &cpu_amount},
				cli.StringFlag{Name: "memory", Value: "128", Destination: &memory_amount},
				cli.StringFlag{Name: "storage", Value: "0", Destination: &storage_amount},
			},
			Action: func(c *cli.Context) error {
				var namespace = c.Args().First()
				if namespace == "" {
					fmt.Println("You have to set a namespace name")
					return nil
				}

				var url = api_base+"/namespaces"

				request := gorequest.New()
				resp, _ , _ := request.Post(url).
					Set("Authorization", "Bearer "+token).
					Send(`{"namespace_name":"` + namespace + `", "resources": { "cpu":` + cpu_amount + `, "mem":` + memory_amount + `, "disks_size":` + storage_amount + `} }`).
					End()

				if httpCodeCheck(resp) {
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
				if namespace == "" {
					fmt.Println("You need to select a namespace")
					return nil
				}
				
				var url = api_base+"/namespaces/"+namespace
				request := gorequest.New()
				resp, _ , _ := request.Delete(url).
					Set("Authorization", "Bearer "+token).
					End()

				if httpCodeCheck(resp) {
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

func httpCodeCheck(response gorequest.Response) bool {
	if debug {
		fmt.Println("HTTP URL:")
		fmt.Println(response.Request.URL)
		fmt.Println("HTTP Code:")
		fmt.Println(response.StatusCode)
		fmt.Println("HTTP Request Headers:")
		fmt.Println(response.Request.Header)

		fmt.Println("HTTP Request:")
		buf := new(bytes.Buffer)
    	buf.ReadFrom(response.Request.Body)
		fmt.Println(buf.String())

		fmt.Println("HTTP Response Headers:")
		fmt.Println(response.Header)
		fmt.Println("HTTP Response:")
		fmt.Println(response.Body)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	jsonResponse := buf.String()

	error := httpError{}
	json.Unmarshal([]byte(jsonResponse), &error)

	if response.StatusCode == 500 {
		fmt.Println("Unkown error")
		return false
	} else if response.StatusCode >= 300 {
		fmt.Println(error.Error)
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
