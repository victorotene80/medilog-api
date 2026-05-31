package bulksms

type BulkSmsRequest struct {
	To   string
	From string
	Body string
}