package application

type RequestOTPResult struct {
	MFARequired bool
	MFAToken    string
	MFAChannel  string
	LoginData   map[string]any
}
