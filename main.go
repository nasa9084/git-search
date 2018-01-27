package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"text/tabwriter"

	flags "github.com/jessevdk/go-flags"
)

type options struct {
	GitHubAPIURI  string `long:"github-api-uri" default:"https://api.github.com/search" default-mask:"-" env:"GITHUB_API_URI" description:"URI for GitHub search API, without trailing slash"`
	Created       string `short:"c" long:"created" description:"created date filter (format: yyyy-mm-dd)"`
	Pushed        string `short:"p" long:"pushed" description:"last updated date filter (format: yyyy-mm-dd)"`
	Fork          string `short:"f" long:"fork" choice:"true" choice:"only" description:"forked repository filter ('true' includes forked repositories, 'only' returns ONLY forked repositories)"`
	Forks         int    `short:"F" long:"forks" default:"-1" default-mask:"-" description:"filter by the number of forks"`
	InName        bool   `short:"N" long:"in-name" description:"restrict to search just repository name"`
	InDescription bool   `short:"D" long:"in-description" description:"restrict to search just repository description"`
	InReadme      bool   `short:"R" long:"in-readme" description:"restrict to search just readme"`
	Language      string `short:"l" long:"language" description:"filter by language they're written in"`
	License       string `short:"L" long:"license" description:"filter by license (user license-keywords in GitHub)"`
	Org           string `short:"o" long:"org" description:"organization which you want to search"`
	Username      string `short:"u" long:"user" description:"username which you want to search"`
	Size          int    `short:"S" long:"size" description:"finds matched given size (in KB)"`
	Stars         int    `short:"s" long:"stars" description:"filter by the number of stars"`
	Topic         string `short:"t" long:"topic" description:"filter by the specific topic"`
	Archived      bool   `short:"a" long:"archived" description:"specify should archived repositories be included"`
	Sort          string `long:"sort" choice:"stars" choice:"forks" choice:"updated" description:"sort field (default:bestmatch)"`
	Order         string `long:"order" choice:"asc" choice:"desc" description:"sort order if sort field is given (default:desc)"`
	PosArgs       struct {
		Keyword string `positional-arg-name:"KEYWORD"`
	} `positional-args:"true"`
}

type response struct {
	Items []struct {
		License     string `json:"omitempty"`
		Language    string
		FullName    string `json:"full_name"`
		Description string
	} `json:"items"`
}

func main() { os.Exit(exec()) }

func exec() int {
	var opts options
	if _, err := flags.Parse(&opts); err != nil {
		if fe, ok := err.(*flags.Error); ok && fe.Type == flags.ErrHelp {
			return 0
		}
		printErr("error in parsing options", err)
		return 1
	}
	uri := opts.GitHubAPIURI + "/repositories?" + opts.queryString()
	res, err := http.Get(uri)
	if err != nil {
		printErr("error in requesting", err)
		return 1
	}
	var r response
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		printErr("error in parsing reponse", err)
		return 1
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	for _, item := range r.Items {
		if opts.Language == "" {
			fmt.Fprintf(w, "[%s]\t", shortenLanguage(item.Language))
		}
		fmt.Fprintln(w, item.FullName+"\t"+item.Description)
	}
	w.Flush()
	return 0
}

func printErr(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err)
}

func shortenLanguage(lang string) string {
	switch lang {
	case "Emacs Lisp":
		return "elisp"
	case "JavaScript":
		return "JS"
	case "CoffeeScript":
		return "Coffee"
	default:
		return lang
	}
}

func (opts options) queryString() string {
	vals := map[string]string{}
	if opts.Created != "" {
		vals["created"] = opts.Created
	}
	if opts.Pushed != "" {
		vals["pushed"] = opts.Pushed
	}
	if opts.Fork != "" {
		vals["fork"] = opts.Fork
	}
	if opts.Forks >= 0 {
		vals["forks"] = strconv.Itoa(opts.Forks)
	}
	if opts.InName {
		vals["in"] = "name"
	}
	if opts.InDescription {
		if _, ok := vals["in"]; ok {
			vals["in"] += ",description"
		} else {
			vals["in"] = "description"
		}
	}
	if opts.InReadme {
		if _, ok := vals["in"]; ok {
			vals["in"] += ",readme"
		} else {
			vals["in"] = "readme"
		}
	}
	if opts.Language != "" {
		vals["language"] = opts.Language
	}
	if opts.License != "" {
		vals["license"] = opts.License
	}
	if opts.Org != "" {
		vals["org"] = opts.Org
	}
	if opts.Username != "" {
		vals["user"] = opts.Username
	}
	if opts.Size > 0 {
		vals["size"] = strconv.Itoa(opts.Size)
	}
	if opts.Stars > 0 {
		vals["stars"] = strconv.Itoa(opts.Stars)
	}
	if opts.Topic != "" {
		vals["topic"] = opts.Topic
	}
	if opts.Archived {
		vals["archived"] = "true"
	}
	buf := bytes.Buffer{}
	isFirst := true
	if opts.PosArgs.Keyword != "" || len(vals) != 0 {
		buf.WriteString("q=")
	}
	if opts.PosArgs.Keyword != "" {
		buf.WriteString(opts.PosArgs.Keyword)
		isFirst = false
	}
	for k, v := range vals {
		if !isFirst {
			buf.WriteString("+")
		}
		buf.WriteString(k + ":" + v)
		isFirst = false
	}
	if opts.Sort != "" {
		if buf.Len() != 0 {
			buf.WriteString("&")
		}
		buf.WriteString("sort=" + opts.Sort)
		if opts.Order != "" {
			buf.WriteString("&order=" + opts.Order)
		}
	}
	return buf.String()
}
