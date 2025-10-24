// package saml

// import (
// 	"context"
// 	"crypto/rsa"
// 	"crypto/tls"
// 	"crypto/x509"
// 	"fmt"
// 	"net/http"
// 	"net/url"

// 	"github.com/crewjam/saml"
// 	"github.com/crewjam/saml/samlsp"
// )

// // SAMLProvider implements SAML SSO authentication
// type SAMLProvider struct {
// 	sp *samlsp.Middleware
// }

// // Config holds SAML configuration
// type Config struct {
// 	EntityID          string
// 	SSOURL            string
// 	IDPMetadataURL    string
// 	CertificateFile   string
// 	PrivateKeyFile    string
// 	RootURL           string
// 	AllowIDPInitiated bool
// }

// // NewSAMLProvider creates a new SAML provider
// func NewSAMLProvider(config Config) (*SAMLProvider, error) {
// 	// Parse root URL
// 	rootURL, err := url.Parse(config.RootURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid root URL: %w", err)
// 	}

// 	// Load IDP metadata
// 	idpMetadataURL, err := url.Parse(config.IDPMetadataURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid IDP metadata URL: %w", err)
// 	}

// 	httpClient := &http.Client{
// 		Transport: &http.Transport{
// 			TLSClientConfig: &tls.Config{
// 				InsecureSkipVerify: false, // Set to true only for development
// 			},
// 		},
// 	}

// 	idpMetadata, err := samlsp.FetchMetadata(context.Background(), httpClient, *idpMetadataURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch IDP metadata: %w", err)
// 	}

// 	// Load certificate and private key
// 	keyPair, err := tls.LoadX509KeyPair(config.CertificateFile, config.PrivateKeyFile)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to load key pair: %w", err)
// 	}

// 	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse certificate: %w", err)
// 	}

// 	// Create SAML service provider
// 	samlSP, err := samlsp.New(samlsp.Options{
// 		EntityID:          config.EntityID,
// 		URL:               *rootURL,
// 		Key:               keyPair.PrivateKey.(*rsa.PrivateKey),
// 		Certificate:       keyPair.Leaf,
// 		IDPMetadata:       idpMetadata,
// 		AllowIDPInitiated: config.AllowIDPInitiated,
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create SAML SP: %w", err)
// 	}

// 	return &SAMLProvider{
// 		sp: samlSP,
// 	}, nil
// }

// // ValidateSAMLResponse validates a SAML response and extracts user information
// func (p *SAMLProvider) ValidateSAMLResponse(ctx context.Context, samlResponse string) (email string, attributes map[string][]string, err error) {
// 	// Parse SAML response
// 	assertion, err := p.sp.ServiceProvider.ParseResponse(nil, []byte(samlResponse))
// 	if err != nil {
// 		return "", nil, fmt.Errorf("failed to parse SAML response: %w", err)
// 	}

// 	// Extract email from NameID or attributes
// 	email = assertion.Subject.NameID.Value

// 	// Extract all attributes
// 	attributes = make(map[string][]string)
// 	for _, attributeStatement := range assertion.AttributeStatements {
// 		for _, attr := range attributeStatement.Attributes {
// 			values := make([]string, 0, len(attr.Values))
// 			for _, value := range attr.Values {
// 				values = append(values, value.Value)
// 			}
// 			attributes[attr.Name] = values
// 		}
// 	}

// 	// Try to get email from attributes if not in NameID
// 	if email == "" {
// 		if emailAttr, ok := attributes["email"]; ok && len(emailAttr) > 0 {
// 			email = emailAttr[0]
// 		} else if emailAttr, ok := attributes["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"]; ok && len(emailAttr) > 0 {
// 			email = emailAttr[0]
// 		} else {
// 			return "", nil, fmt.Errorf("email not found in SAML response")
// 		}
// 	}

// 	return email, attributes, nil
// }

// // GetSAMLAuthURL returns the SAML SSO authentication URL
// func (p *SAMLProvider) GetSAMLAuthURL(ctx context.Context, relayState string) (string, error) {
// 	// Build authentication request
// 	authReq, err := p.sp.ServiceProvider.MakeAuthenticationRequest(
// 		p.sp.ServiceProvider.GetSSOBindingLocation(saml.HTTPRedirectBinding),
// 		saml.HTTPRedirectBinding,
// 		saml.HTTPPostBinding,
// 	)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to create authentication request: %w", err)
// 	}

// 	// Add relay state if provided
// 	if relayState != "" {
// 		authReq.RelayState = relayState
// 	}

// 	// Generate redirect URL
// 	redirectURL, err := authReq.Redirect(relayState, &p.sp.ServiceProvider)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to generate redirect URL: %w", err)
// 	}

// 	return redirectURL.String(), nil
// }

// // GetServiceProvider returns the underlying SAML service provider
// func (p *SAMLProvider) GetServiceProvider() *saml.ServiceProvider {
// 	return &p.sp.ServiceProvider
// }

// // GetMetadataXML returns the SP metadata XML
// func (p *SAMLProvider) GetMetadataXML() ([]byte, error) {
// 	metadata := p.sp.ServiceProvider.Metadata()
// 	return metadata.MarshalTo(nil)
// }