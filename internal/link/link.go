package link

type Link struct {
	OriginalURL string
	Code        string
}

const CreateLinkMaxAttempts = 5
const CodeLength = 6
const maxURLlength = 200
const maxActiveLinksForIP = 10

func CanCreate(url string, ipCount int) error {
	//TODO: validate and test if the url is a valid URL.
	if len(url) > maxURLlength {
		return ErrTooLongURL
	}
	if ipCount >= maxActiveLinksForIP {
		return ErrTooManyActiveURLs
	}
	return nil
}
