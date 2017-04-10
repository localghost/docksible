package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"

	"golang.org/x/crypto/ssh"
)

type SSHKeyGenerator struct{}

type InMemorySSHKeys struct {
	PrivateKey *bytes.Buffer
	PublicKey  *bytes.Buffer
}

func NewSSHKeyGenerator() *SSHKeyGenerator {
	return &SSHKeyGenerator{}
}

func NewInMemorySSHKeys() *InMemorySSHKeys {
	return &InMemorySSHKeys{
		PrivateKey: new(bytes.Buffer),
		PublicKey:  new(bytes.Buffer),
	}
}

func (s *SSHKeyGenerator) GenerateInMemory() *InMemorySSHKeys {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		log.Fatal(err)
	}

	sshKeys := NewInMemorySSHKeys()
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err := pem.Encode(sshKeys.PrivateKey, privateKeyPEM); err != nil {
		log.Fatal(err)
	}

	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err)
	}
	sshKeys.PublicKey.Write(ssh.MarshalAuthorizedKey(pub))

	return sshKeys
}

// // MakeSSHKeyPair make a pair of public and private keys for SSH access.
// // Public key is encoded in the format for inclusion in an OpenSSH authorized_keys file.
// // Private Key generated is PEM encoded
// func MakeSSHKeyPair(pubKeyPath, privateKeyPath string) error {
// 	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
// 	if err != nil {
// 		return err
// 	}

// 	// generate and write private key as PEM
// 	privateKeyFile, err := os.Create(privateKeyPath)
// 	defer privateKeyFile.Close()
// 	if err != nil {
// 		return err
// 	}
// 	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
// 	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
// 		return err
// 	}

// 	// generate and write public key
// 	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
// 	if err != nil {
// 		return err
// 	}
// 	return ioutil.WriteFile(pubKeyPath, ssh.MarshalAuthorizedKey(pub), 0655)
// }
