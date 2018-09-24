package goqldom

import (
	"bytes"
	"errors"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/graphql-go/graphql"
	"golang.org/x/net/html"
)

func text(elements string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(elements))
	token := tokenizer.Token()

	var texts []string
	var end bool
	for !end {
		next := tokenizer.Next()
		switch {
		case next == html.ErrorToken:
			end = true
		case next == html.StartTagToken:
			token = tokenizer.Token()
		case next == html.TextToken:
			if token.Data == "script" {
				continue
			}
			content := strings.TrimSpace(html.UnescapeString(string(tokenizer.Text())))
			if len(content) > 0 {
				texts = append(texts, content)
			}
		}
	}

	return strings.Join(texts, "\n")
}

func this(selection *goquery.Selection, selector string) *goquery.Selection {
	if selector != "" {
		return selection.Find(selector)
	}
	return selection
}

type Selectable interface {
	Select(selector string) *Elements
}

type Node interface {
	Attr(selector string, key string) string
	HTML(selector string) (string, error)
	Text(selector string) (string, error)
}

type Element struct {
	selection *goquery.Selection
}

func (e *Element) Attr(selector string, key string) string {
	if val, ok := this(e.selection, selector).Attr(key); ok {
		return val
	}
	return ""
}

func (e *Element) HTML(selector string) (string, error) {
	if val, err := this(e.selection, selector).Html(); err != nil {
		return "", err
	} else {
		return strings.TrimSpace(val), nil
	}
}

func (e *Element) Text(selector string) (string, error) {
	if content, err := this(e.selection, selector).Html(); err != nil {
		return "", err
	} else {
		return text(content), nil
	}
}

func (e *Element) Select(selector string) *Elements {
	return &Elements{Element{selection: e.selection.Find(selector)}}
}

type Elements struct {
	Element
}

func (e *Elements) Elements(selector string) []*Element {
	var list []*Element
	e.selection.Each(func(index int, selection *goquery.Selection) {
		if selector == "" || selection.Is(selector) || selection.Has(selector).Length() > 0 {
			list = append(list, &Element{selection: selection})
		}
	})
	return list
}

type Document struct {
	document *goquery.Document
	Location string
}

func (d *Document) Title() string {
	return strings.TrimSpace(d.document.Find("head title").Text())
}

func (d *Document) Select(selector string) *Elements {
	return &Elements{Element{selection: d.document.Find(selector)}}
}

type Response struct {
	StatusCode    int
	StatusMessage string
	ContentType   string
	Document      *Document
}

func ErrorResponse(message string, err error) (*Response, error) {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewBufferString(err.Error()))
	return &Response{
		StatusCode:    http.StatusInternalServerError,
		StatusMessage: message,
		ContentType:   "text/plain; charset=utf-8",
		Document:      &Document{document: doc},
	}, err
}

func GetResponse(url string) (*Response, error) {
	if res, err := http.Get(url); err != nil {
		return ErrorResponse(url+" is not accessible", err)
	} else if doc, err := goquery.NewDocumentFromReader(res.Body); err != nil {
		res.Body.Close()
		return ErrorResponse(url+" response is not readable as HTML", err)
	} else {
		return &Response{
			StatusCode:    res.StatusCode,
			StatusMessage: res.Status,
			ContentType:   res.Header.Get("Content-Type"),
			Document: &Document{
				document: doc,
				Location: url,
			},
		}, nil
	}
}

type Resolver func(selector string, node Node) (interface{}, error)

func optional(params graphql.ResolveParams, resolve Resolver) (interface{}, error) {
	selector, ok := params.Args["selector"].(string)
	if !ok {
		selector = ""
	}
	if element, ok := params.Source.(Node); !ok {
		return nil, errors.New("invalid state")
	} else {
		return resolve(selector, element)
	}
}

func Schema(version string) (graphql.Schema, error) {

	optionalSelectorArgs := graphql.FieldConfigArgument{
		"selector": &graphql.ArgumentConfig{
			Description: "CSS selector, see: http://butlerccwebdev.net/support/css-selectors-cheatsheet.html",
			Type:        graphql.String,
		},
	}

	nonNullSelectorArgs := graphql.FieldConfigArgument{
		"selector": &graphql.ArgumentConfig{
			Description: "CSS selector, see: http://butlerccwebdev.net/support/css-selectors-cheatsheet.html",
			Type:        graphql.NewNonNull(graphql.String),
		},
	}

	attrField := &graphql.Field{
		Description: "Attribute value of the selection.",
		Type:        graphql.NewNonNull(graphql.String),
		Args: graphql.FieldConfigArgument{
			"selector": optionalSelectorArgs["selector"],
			"key": &graphql.ArgumentConfig{
				Description: "Key of the attribute value.",
				Type:        graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			return optional(params, func(selector string, node Node) (interface{}, error) {
				if key, ok := params.Args["key"].(string); !ok {
					return nil, errors.New("missing key param")
				} else {
					return node.Attr(selector, key), nil
				}
			})
		},
	}

	htmlField := &graphql.Field{
		Description: "Inner HTML of the selection as text.",
		Type:        graphql.NewNonNull(graphql.String),
		Args:        optionalSelectorArgs,
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			return optional(params, func(selector string, node Node) (interface{}, error) {
				return node.HTML(selector)
			})
		},
	}

	textField := &graphql.Field{
		Description: "Concatenated text.",
		Type:        graphql.NewNonNull(graphql.String),
		Args:        optionalSelectorArgs,
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			return optional(params, func(selector string, node Node) (interface{}, error) {
				return node.Text(selector)
			})
		},
	}

	elementType := graphql.NewObject(
		graphql.ObjectConfig{
			Description: "HTML element.",
			Name:        "Element",
			Fields: graphql.Fields{
				"attr": attrField,
				"html": htmlField,
				"text": textField,
			},
		},
	)

	elementsType := graphql.NewObject(
		graphql.ObjectConfig{
			Description: "HTML elements.",
			Name:        "Elements",
			Fields: graphql.Fields{
				"attr": attrField,
				"html": htmlField,
				"text": textField,
			},
		},
	)

	selectField := &graphql.Field{
		Description: "Select elements.",
		Type:        elementsType,
		Args:        nonNullSelectorArgs,
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			if selector, ok := params.Args["selector"].(string); !ok {
				return nil, errors.New("missing selector param")
			} else {
				if element, ok := params.Source.(Selectable); !ok {
					return nil, errors.New("invalid state: select")
				} else {
					return element.Select(selector), nil
				}
			}
		},
	}

	elementsField := &graphql.Field{
		Description: "List all selected elements.",
		Type:        graphql.NewList(elementType),
		Args:        optionalSelectorArgs,
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			selector, ok := params.Args["selector"].(string)
			if !ok {
				selector = ""
			}
			if elements, ok := params.Source.(*Elements); !ok {
				return nil, errors.New("invalid state: elements")
			} else {
				return elements.Elements(selector), nil
			}
		},
	}

	elementType.AddFieldConfig("select", selectField)
	elementsType.AddFieldConfig("select", selectField)
	elementsType.AddFieldConfig("elements", elementsField)

	locationField := &graphql.Field{
		Description: "URL of this Document.",
		Type:        graphql.NewNonNull(graphql.String),
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			if document, ok := params.Source.(*Document); !ok {
				return nil, errors.New("invalid state: location")
			} else {
				return document.Location, nil
			}
		},
	}

	titleField := &graphql.Field{
		Description: "Text of the DOM title element.",
		Type:        graphql.NewNonNull(graphql.String),
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			if document, ok := params.Source.(*Document); !ok {
				return nil, errors.New("invalid state: title")
			} else {
				return document.Title(), nil
			}
		},
	}

	documentType := graphql.NewObject(
		graphql.ObjectConfig{
			Description: "HTML document as the root for selections.",
			Name:        "Document",
			Fields: graphql.Fields{
				"location": locationField,
				"title":    titleField,
				"select":   selectField,
			},
		},
	)

	statusCodeField := &graphql.Field{
		Description: "HTTP status code.",
		Type:        graphql.NewNonNull(graphql.Int),
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			if response, ok := params.Source.(*Response); ok {
				return response.StatusCode, nil
			}
			return nil, nil
		},
	}

	statusMessageField := &graphql.Field{
		Description: "HTTP status message.",
		Type:        graphql.NewNonNull(graphql.String),
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			if response, ok := params.Source.(*Response); ok {
				return response.StatusMessage, nil
			}
			return nil, nil
		},
	}

	contentTypeField := &graphql.Field{
		Description: "Content type.",
		Type:        graphql.NewNonNull(graphql.String),
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			if response, ok := params.Source.(*Response); ok {
				return response.ContentType, nil
			}
			return nil, nil
		},
	}

	documentField := &graphql.Field{
		Description: "Get the DOM for selections.",
		Type:        documentType,
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			if response, ok := params.Source.(*Response); ok {
				return response.Document, nil
			}
			return nil, nil
		},
	}

	responseType := graphql.NewObject(
		graphql.ObjectConfig{
			Description: "HTTP response.",
			Name:        "Response",
			Fields: graphql.Fields{
				"statusCode":    statusCodeField,
				"statusMessage": statusMessageField,
				"contentType":   contentTypeField,
				"document":      documentField,
			},
		},
	)

	getField := &graphql.Field{
		Description: "Connect a URL to fetch the root Document.",
		Type:        responseType,
		Args: graphql.FieldConfigArgument{
			"url": &graphql.ArgumentConfig{
				Description: "The URL to connect to.",
				Type:        graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			if url, ok := params.Args["url"].(string); !ok {
				return nil, errors.New("missing URL param")
			} else {
				return GetResponse(url)
			}
		},
	}

	versionField := &graphql.Field{
		Description: "The version of this service.",
		Type:        graphql.NewNonNull(graphql.String),
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			return version, nil
		},
	}

	return graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Description: "A GraphQL based HTML binding for arbitrary DOM selections.",
			Name:        "Query",
			Fields: graphql.Fields{
				"get":     getField,
				"version": versionField,
			},
		}),
	})
}
