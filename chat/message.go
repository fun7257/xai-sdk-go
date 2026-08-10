package chat

import (
	"fmt"

	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

func normalizeParts(parts ...any) ([]*xaiv1.Content, error) {
	out := make([]*xaiv1.Content, 0, len(parts))
	for _, p := range parts {
		switch v := p.(type) {
		case string:
			out = append(out, Text(v))
		case *xaiv1.Content:
			out = append(out, v)
		default:
			return nil, fmt.Errorf("unsupported content part %T", p)
		}
	}
	return out, nil
}

// Text builds text content.
func Text(s string) *xaiv1.Content {
	return &xaiv1.Content{Content: &xaiv1.Content_Text{Text: s}}
}

// Image builds image URL or base64 data-URL content. detail is auto|low|high.
// The image_url field accepts a public URL or a data: URL (base64) as in the Python SDK (CHAT-09).
func Image(url string, detail string) *xaiv1.Content {
	d := xaiv1.ImageDetail_DETAIL_AUTO
	switch detail {
	case "low":
		d = xaiv1.ImageDetail_DETAIL_LOW
	case "high":
		d = xaiv1.ImageDetail_DETAIL_HIGH
	}
	return &xaiv1.Content{Content: &xaiv1.Content_ImageUrl{ImageUrl: &xaiv1.ImageUrlContent{
		Source: &xaiv1.ImageUrlContent_ImageUrl{ImageUrl: url},
		Detail: d,
	}}}
}

// FileByID references a Files API file.
func FileByID(id string) *xaiv1.Content {
	return &xaiv1.Content{Content: &xaiv1.Content_File{File: &xaiv1.FileContent{FileId: id}}}
}

// FileByData attaches inline bytes.
func FileByData(data []byte, filename, mime string) *xaiv1.Content {
	return &xaiv1.Content{Content: &xaiv1.Content_File{File: &xaiv1.FileContent{
		Data: data, Filename: filename, MimeType: mime,
	}}}
}

// FileByURL attaches a public file URL.
func FileByURL(url, filename, mime string) *xaiv1.Content {
	return &xaiv1.Content{Content: &xaiv1.Content_File{File: &xaiv1.FileContent{
		Url: url, Filename: filename, MimeType: mime,
	}}}
}

func message(role xaiv1.MessageRole, parts ...any) (*xaiv1.Message, error) {
	c, err := normalizeParts(parts...)
	if err != nil {
		return nil, err
	}
	return &xaiv1.Message{Role: role, Content: c}, nil
}

// mustMessage panics if err is non-nil; used by convenience factories for valid string/*Content parts.
func mustMessage(m *xaiv1.Message, err error) *xaiv1.Message {
	if err != nil {
		panic(err)
	}
	return m
}

// NewUser builds a user message or returns an error for unsupported parts.
func NewUser(parts ...any) (*xaiv1.Message, error) {
	return message(xaiv1.MessageRole_ROLE_USER, parts...)
}

// NewSystem builds a system message or returns an error for unsupported parts.
func NewSystem(parts ...any) (*xaiv1.Message, error) {
	return message(xaiv1.MessageRole_ROLE_SYSTEM, parts...)
}

// NewAssistant builds an assistant message or returns an error for unsupported parts.
func NewAssistant(parts ...any) (*xaiv1.Message, error) {
	return message(xaiv1.MessageRole_ROLE_ASSISTANT, parts...)
}

// NewDeveloper builds a developer message or returns an error for unsupported parts.
func NewDeveloper(parts ...any) (*xaiv1.Message, error) {
	return message(xaiv1.MessageRole_ROLE_DEVELOPER, parts...)
}

// User builds a user message from string or *Content parts.
//
// Panic contract: User panics if any part type is unsupported (not string and
// not *xaiv1.Content). Prefer this for fixed literal parts at call sites.
// For dynamic or untrusted part values, use [NewUser], which returns an error
// instead of panicking.
func User(parts ...any) *xaiv1.Message {
	return mustMessage(message(xaiv1.MessageRole_ROLE_USER, parts...))
}

// System builds a system message from string or *Content parts.
//
// Panic contract: System panics on unsupported part types. Use [NewSystem]
// when parts may be invalid and you need an error return.
func System(parts ...any) *xaiv1.Message {
	return mustMessage(message(xaiv1.MessageRole_ROLE_SYSTEM, parts...))
}

// Assistant builds an assistant message from string or *Content parts.
//
// Panic contract: Assistant panics on unsupported part types. Use [NewAssistant]
// for an error-returning path.
func Assistant(parts ...any) *xaiv1.Message {
	return mustMessage(message(xaiv1.MessageRole_ROLE_ASSISTANT, parts...))
}

// Developer builds a developer message from string or *Content parts.
//
// Panic contract: Developer panics on unsupported part types. Use [NewDeveloper]
// for an error-returning path.
func Developer(parts ...any) *xaiv1.Message {
	return mustMessage(message(xaiv1.MessageRole_ROLE_DEVELOPER, parts...))
}

// ToolResult builds a tool-role message.
func ToolResult(result string, toolCallID string) *xaiv1.Message {
	m := &xaiv1.Message{
		Role:    xaiv1.MessageRole_ROLE_TOOL,
		Content: []*xaiv1.Content{Text(result)},
	}
	if toolCallID != "" {
		id := toolCallID
		m.ToolCallId = &id
	}
	return m
}

// File builds file content with exactly one of fileID, data, or url (CHAT-10).
// Filename/mime apply only to data/url modes.
func File(fileID string, data []byte, url, filename, mime string) (*xaiv1.Content, error) {
	n := 0
	if fileID != "" {
		n++
	}
	if data != nil {
		n++
	}
	if url != "" {
		n++
	}
	if n != 1 {
		return nil, fmt.Errorf("file content: exactly one of file_id, data, or url must be set")
	}
	if fileID != "" {
		if filename != "" || mime != "" {
			return nil, fmt.Errorf("file content: filename/mime only supported for data or url modes")
		}
		return FileByID(fileID), nil
	}
	if url != "" {
		return FileByURL(url, filename, mime), nil
	}
	return FileByData(data, filename, mime), nil
}
