// Copyright 2021 Matthew Garrett <mgarrett@aurora.tech>
//
// This software is released under the terms of the Apache License 2.0. A copy
// can be found in LICENSE.md in this repository.

package main

import (
	"crypto/x509"
	"encoding/base64"
	"flag"
	"fmt"
	"os"	

	"golang.org/x/crypto/ssh"
)

func main() {
	certificate := flag.String("certificate", "", "Certificate to validate")
	user := flag.String("user", "", "Requested user")
	rootCA := flag.String("rootCA", "", "Path of the root CA to validate the certificate")
	
	flag.Parse()

	if *certificate == "" || *user == "" {
		flag.Usage()
		os.Exit(1)
	}

	fullCert := fmt.Sprintf("ssh-rsa-cert-v01@openssh.com %s", *certificate)
	pub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(fullCert))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse certificate: %v\n", err)
		os.Exit(1)
	}

	sshCert, ok := pub.(*ssh.Certificate)
	if !ok {
		fmt.Fprintf(os.Stderr, "Failed to cast to certificate\n")
		os.Exit(1)
	}

	encodedCert, ok := sshCert.Permissions.Extensions["x509"]
	if !ok {
		fmt.Fprintf(os.Stderr, "Certificate doesn't contain valid x509\n")
		os.Exit(1)
	}

	decodedCert, err := base64.StdEncoding.DecodeString(encodedCert)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to decode X509 certificate: %v\n", err)
		os.Exit(1)
	}
	
	cert, err := x509.ParseCertificate(decodedCert)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse X509 certificate: %v\n", err)
		os.Exit(1)
	}

	if *rootCA != "" {
		ca, err := os.ReadFile(*rootCA)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read root CA: %v\n", err)
			os.Exit(1)
		}
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM(ca)
		if !ok {
			fmt.Fprintf(os.Stderr, "Failed to append root CA to pool\n")
			os.Exit(1)
		}
		opts := x509.VerifyOptions{
			Roots: roots,
		}
		if _, err := cert.Verify(opts); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to verify X509 certificate\n")
			os.Exit(1)
		}
	}

	subjectCN := cert.Subject.CommonName
	if subjectCN != *user {
		fmt.Fprintf(os.Stderr, "Certificate does not match requested user name: %s, %s\n", subjectCN, *user)
		os.Exit(1)
	}

	sshKey, err := ssh.NewPublicKey(cert.PublicKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to obtain SSH public key from X509 public key: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("cert-authority %s", ssh.MarshalAuthorizedKey(sshKey))
}
