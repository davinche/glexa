package glexa

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// VerifyAlexaRequest authenticates whether the incoming request is from AWS
func VerifyRequest(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sigCertChainURL := r.Header.Get("SignatureCertChainUrl")
		// Check for valid sig chain url
		if err := verifyCertURL(sigCertChainURL); err != nil {
			log.Printf("error: invalid SignatureCertChainURL: %q\n", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		bodyBuf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("error: could not read request body: %q\n", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		// validate timestamp
		if err := verifyBodyTimestamp(bytes.NewBuffer(bodyBuf)); err != nil {
			log.Printf("error: invalid timestamp: %q\n", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		// vaidate certchain
		cert, err := validateCertChain(sigCertChainURL)
		if err != nil {
			log.Printf("error: invalid certificate chain: %q\n", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		// verify signature
		signature := r.Header.Get("Signature")
		err = verifySignature(signature, cert.PublicKey.(*rsa.PublicKey), bytes.NewBuffer(bodyBuf))
		if err != nil {
			log.Printf("error: could not verify signature: %q\n", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		// reset body
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBuf))

		// everything checks out; run the handler
		h.ServeHTTP(w, r)
	}
}

func verifyCertURL(certURL string) error {
	parsed, err := url.Parse(certURL)
	if err != nil {
		return fmt.Errorf("could not parse SignatureCertChainUrl: %q\n", err)
	}

	if parsed.Scheme != "https" {
		return fmt.Errorf("scheme is not https: %q\n", parsed.Scheme)
	}

	if host, port, err := net.SplitHostPort(parsed.Host); err == nil {
		if port != "443" || host != "s3.amazonaws.com" {
			return fmt.Errorf("invalid hostname or port")
		}
	}

	if !strings.HasPrefix(strings.ToLower(parsed.Host), "s3.amazonaws.com") {
		return fmt.Errorf("invalid hostname")
	}

	if !strings.HasPrefix(parsed.Path, "/echo.api/") {
		return fmt.Errorf("invalid path")
	}
	return nil
}

func verifyBodyTimestamp(body io.Reader) error {
	// check that body is not expired
	bodyStruct := struct {
		Request struct {
			Timestamp string `json:"timestamp"`
		} `json:"request"`
	}{}

	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&bodyStruct); err != nil {
		return fmt.Errorf("could not decode body: %q\n", err)
	}

	requestTime, err := time.Parse("2006-01-02T15:04:05Z", bodyStruct.Request.Timestamp)
	if err != nil {
		return fmt.Errorf("could not parse timestamp: %q\n", err)
	}

	if time.Now().Sub(requestTime) > time.Second*150 {
		return fmt.Errorf("timestamp is stale")
	}
	return nil
}

func validateCertChain(chainURL string) (*x509.Certificate, error) {
	resp, err := http.Get(chainURL)
	if err != nil {
		return nil, fmt.Errorf("could not get cert chain pem: %q\n", err)
	}

	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read cert chain pem: %q\n", err)
	}

	block, _ := pem.Decode(buf)
	if err != nil {
		return nil, fmt.Errorf("could not decode cert chain: %q\n", err)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("could not parse cert chain: %q\n", err)
	}

	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM(buf); !ok {
		return nil, fmt.Errorf("could not parse root cert: %q\n", err)
	}

	opts := x509.VerifyOptions{
		DNSName: "echo-api.amazon.com",
		Roots:   roots,
	}

	if _, err := cert.Verify(opts); err != nil {
		return nil, fmt.Errorf("could not verify cert chain: %q\n", err)
	}
	return cert, nil
}

func verifySignature(signature string, pubKey *rsa.PublicKey, body io.Reader) error {
	data, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("could not base64 decode signature: %q\n", err)
	}
	buf, err := ioutil.ReadAll(body)
	if err != nil {
		return fmt.Errorf("could not read request body: %q\n", err)
	}
	hashed := sha1.Sum(buf)
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA1, hashed[:], data); err != nil {
		return fmt.Errorf("verification error: %q\n", err)
	}
	return nil
}
