package main

import (
	"fmt"
	"github.com/elixir-oslo/crypt4gh/keys"
	"github.com/jessevdk/go-flags"
	"github.com/logrusorgru/aurora"
	"github.com/manifoldco/promptui"
	"golang.org/x/crypto/chacha20poly1305"
	"log"
	"os"
)

const (
	generate = "generate"
	encrypt  = "encrypt"
	decrypt  = "decrypt"
	version  = "version"
)

var generateOptions struct {
	Name string `short:"n" long:"name" description:"Key pair name" required:"true"`
}

var generateOptionsParser = flags.NewParser(&generateOptions, flags.None)

var encryptOptions struct {
	FileName          string `short:"f"  long:"file" description:"File to encrypt" value-name:"FILE"`
	SecretKeyFileName string `short:"s" long:"seckey" description:"Secret key to use" value-name:"FILE" required:"true"`
	PublicKeyFileName string `short:"p" long:"pubkey" description:"Public key to use" value-name:"FILE" required:"true"`
}
var encryptOptionsParser = flags.NewParser(&encryptOptions, flags.None)

var decryptOptions struct {
	FileName          string `short:"f" long:"file" description:"File to decrypt" value-name:"FILE" required:"true"`
	SecretKeyFileName string `short:"s" long:"seckey" description:"Secret key to use" value-name:"FILE" required:"true"`
}

var decryptOptionsParser = flags.NewParser(&decryptOptions, flags.None)

func main() {
	args := os.Args
	if len(args) == 1 || args[1] == "-h" || args[1] == "--help" {
		generateOptionsParser.WriteHelp(os.Stdout)
		encryptOptionsParser.WriteHelp(os.Stdout)
		decryptOptionsParser.WriteHelp(os.Stdout)
		os.Exit(0)
	}
	commandName := args[1]
	switch commandName {
	case generate:
		_, err := generateOptionsParser.Parse()
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
		publicKey, privateKey, err := keys.GenerateKeyPair()
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
		err = writeKeyPair(generateOptions.Name, publicKey, privateKey)
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
	case encrypt:
		_, err := encryptOptionsParser.Parse()

		if err != nil {
			log.Fatal(aurora.Red(err))
		}
	case decrypt:
		_, err := decryptOptionsParser.Parse()
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
	default:
		log.Fatal(aurora.Red(fmt.Sprintf("command '%v' is not recognized", commandName)))
	}
}

func writeKeyPair(name string, publicKey [chacha20poly1305.KeySize]byte, privateKey [chacha20poly1305.KeySize]byte) error {
	publicKeyFileName := name + ".pub.pem"
	privateKeyFileName := name + ".sec.pem"
	if fileExists(publicKeyFileName) || fileExists(privateKeyFileName) {
		promptYesNo(fmt.Sprintf("Key pair with name '%v' seems to already exist. Please, confirm overwriting", name))
	}
	publicKeyFile, err := os.Create(publicKeyFileName)
	if err != nil {
		return err
	}
	err = keys.WriteOpenSSLX25519PublicKey(publicKeyFile, publicKey)
	if err != nil {
		return err
	}
	err = publicKeyFile.Close()
	if err != nil {
		return err
	}
	privateKeyFile, err := os.Create(privateKeyFileName)
	if err != nil {
		return err
	}
	err = keys.WriteOpenSSLX25519PrivateKey(privateKeyFile, privateKey)
	if err != nil {
		return err
	}
	err = privateKeyFile.Close()
	if err != nil {
		return err
	}
	return nil
}

func promptYesNo(message string) {
	prompt := promptui.Select{
		Label: message,
		Items: []string{"Yes", "No"},
	}
	_, result, err := prompt.Run()
	if err != nil || result != "Yes" {
		os.Exit(0)
	}
}

func promptPassword(minLength int) (password string, err error) {
	validate := func(input string) error {
		if len(input) < minLength {
			return fmt.Errorf("password is too short: min length is %v", minLength)
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    "Enter the password to unlock the key",
		Validate: validate,
	}
	return prompt.Run()
}

func fileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
