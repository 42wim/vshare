package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/h2non/filetype"
)

const maxSize = 400000

var (
	version string
	githash string
)

func isPiped() bool {
	fi, _ := os.Stdin.Stat()
	return (fi.Mode() & os.ModeCharDevice) == 0
}

func checkFileSize(file string) error {
	if file == "" {
		return nil
	}
	fi, err := os.Stat(file)
	if err != nil {
		return err
	}
	if fi.Size() > maxSize {
		return fmt.Errorf("file too big, maximum %d bytes", maxSize)
	}
	return nil
}

// askSecret will return the secret either via file, piped or stdin
func askSecret(file string) (string, error) {
	// read and base64 encode file
	if file != "" {
		res, err := ioutil.ReadFile(file)
		if err != nil {
			return "", err
		}
		secretData := base64.StdEncoding.EncodeToString(res)
		return secretData, nil
	}

	// if we are reading from a pipe base64 encode it
	if isPiped() {
		res, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		if len(res) > maxSize {
			return "", fmt.Errorf("file too big, maximum %d bytes", maxSize)
		}
		secretData := base64.StdEncoding.EncodeToString(res)
		return secretData, nil
	}

	// just reading from stdin for the secret
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your secret to share: ")
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func handleEncodedFile(input string) error {
	data, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return err
	}
	filename := "vshare.output"
	kind, _ := filetype.Match(data)
	if kind != filetype.Unknown {
		filename = "vshare." + kind.Extension
	}

	// just reading from stdin for the secret
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("The secret is a file, enter a filename to safe the output in (default: \"%s\"): ", filename)
	output, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	output = strings.TrimSpace(output)
	if output == "" {
		output = filename
	}
	err = ioutil.WriteFile(output, data, 0600)
	if err != nil {
		return err
	}
	return nil
}

// handleCreate handles vault login, input handling and returns a token
func handleCreateToken(file, ttl string) (string, error) {
	encode := false
	if os.Getenv("VAULT_TOKEN") == "" {
		log.Fatal("no VAULT_TOKEN environment variable set")
	}
	v, err := vaultInit()
	if err != nil {
		return "", err
	}
	secretText, err := askSecret(file)
	if err != nil {
		return "", err
	}
	// set the secret in a cubbyhole with specified ttl
	if file != "" || isPiped() {
		encode = true
	}
	// set the secret in a cubbyhole with specified ttl
	token, err := setCubby(v, secretText, ttl, encode)
	if err != nil {
		return "", err
	}
	return token, nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s: %s\n", os.Args[0], "wraps your secret in a single-use token.")
		if githash == "" {
			version = "dev"
		}
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %s %s\n", version, githash)
		fmt.Fprintln(flag.CommandLine.Output(), "Needs VAULT_ADDR and VAULT_TOKEN env variable.")
		fmt.Fprintln(flag.CommandLine.Output())
		fmt.Fprintf(flag.CommandLine.Output(), "%s \t\t\t:%s\n", os.Args[0], "will ask you for your secret")
		fmt.Fprintf(flag.CommandLine.Output(), "%s %s\n", os.Args[0], "<token>\t\t:will reveal the secret wrapped in <token> ")
		fmt.Fprintln(flag.CommandLine.Output())
		fmt.Fprintln(flag.CommandLine.Output(), "flags:")
		flag.PrintDefaults()
	}
	web := flag.Bool("web", false, "run as webserver. Needs VAULT_ADDR and VSHARE_URL env variable")
	webPort := flag.String("webport", ":80", "<ip>:<port> to listen on as webserver")
	file := flag.String("file", "", "will use the content of <file> as the secret")
	ttl := flag.String("ttl", "15m", "ttl of the token")
	flag.Parse()

	if os.Getenv("VAULT_ADDR") == "" {
		flag.Usage()
		fmt.Println()
		log.Fatal("no VAULT_ADDR environment variable set")
	}

	if *web {
		url := os.Getenv("VSHARE_URL")
		if url == "" {
			log.Fatal("no VSHARE_URL environment variable set")
		}
		startWeb(*webPort)
		return
	}

	// no arguments: encrypt using stdin
	if len(flag.Args()) == 0 {
		err := checkFileSize(*file)
		if err != nil {
			log.Fatal(err)
		}
		token, err := handleCreateToken(*file, *ttl)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("This secret will self destruct when read or after %s.\n", *ttl)
		fmt.Println()
		fmt.Printf("The token to share is %s\n", token)
		url := os.Getenv("VSHARE_URL")
		if url == "" {
			return
		}
		if *file != "" {
			fmt.Printf("Or via web %s/info/%s?file=%s\n", url, token, *file)
		} else {
			fmt.Printf("Or via web %s/info/%s\n", url, token)
		}
		return
	}

	// if we have an argument it should be a token we use to get the secret
	secret, encoded, err := unWrap(flag.Args()[0])
	if err != nil {
		fmt.Printf("Getting secret failed: %s\n", err)
		return
	}

	if encoded {
		// handle the base64 encoded secret
		err = handleEncodedFile(secret)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if !encoded {
		fmt.Printf("The secret is: %s\n", secret)
		return
	}
}
