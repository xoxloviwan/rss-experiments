package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"
)

type AtomFeed struct {
	XMLName  xml.Name `xml:"feed"`
	Xmlns    string   `xml:"xmlns,attr"`
	Title    string   `xml:"title"`   // required
	Id       string   `xml:"id"`      // required
	Updated  string   `xml:"updated"` // required
	Category string   `xml:"category,omitempty"`
	Icon     string   `xml:"icon,omitempty"`
	Logo     string   `xml:"logo,omitempty"`
	Rights   string   `xml:"rights,omitempty"` // copyright used
	Subtitle string   `xml:"subtitle,omitempty"`
	Link     *AtomLink
	Author   *AtomAuthor  `xml:"author,omitempty"`
	Entries  []*AtomEntry `xml:"entry"`
}

// AtomEntry представляет один элемент <entry> из вашего фида
type AtomEntry struct {
	XMLName  xml.Name     `xml:"entry"`
	Title    string       `xml:"title"`
	Updated  string       `xml:"updated"`
	Id       string       `xml:"id"`
	Author   AtomAuthor   `xml:"author"`
	Category AtomCategory `xml:"category"`

	// Критически важно: именно этот тег заставляет Go собирать все теги <link> в массив
	Links []AtomLink `xml:"link"`
}

// AtomLink представляет тег <link> и его атрибуты
type AtomLink struct {
	// Внимание: Поле XMLName xml.Name удалено.
	// Это гарантирует, что ссылки с разным набором атрибутов (с rel и без) распарсятся одинаково успешно.
	Href   string `xml:"href,attr"`
	Rel    string `xml:"rel,attr,omitempty"`
	Type   string `xml:"type,attr,omitempty"`
	Length string `xml:"length,attr,omitempty"`
}

// AtomAuthor представляет тег <author>
type AtomAuthor struct {
	Name string `xml:"name"`
}

// AtomCategory представляет тег <category>
type AtomCategory struct {
	Term  string `xml:"term,attr"`
	Label string `xml:"label,attr,omitempty"`
}
type Author struct {
	Name string `xml:"name"`
}

type PageData struct {
	Title       string
	GeneratedAt string
	Items       []*AtomEntry
}

func main() {
	feedURL := fmt.Sprintf("https://%s/atom/f/%s.atom", os.Getenv("FEED"), os.Getenv("FRM"))

	resp, err := http.Get(feedURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// body, _ := os.ReadFile("0.atom.xml")
	// rdr := bytes.NewReader(body)

	rss := new(AtomFeed)
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&rss); err != nil {
		panic(err)
	}

	// // Ограничим свежими записями (например, первые 20)
	// items := rss.Feed.Items
	// if len(items) > 20 {
	// 	items = items[:20]
	// }

	data := PageData{
		Title:       "Мой RSS Ридер",
		GeneratedAt: time.Now().Format("02.01.2006 15:05"),
		Items:       rss.Entries,
	}

	tmplHTML := `<!DOCTYPE html>
	<html lang="ru">
	<head>
	    <meta charset="UTF-8">
	    <meta name="viewport" content="width=device-width, initial-scale=1.0">
	    <title>{{.Title}}</title>
	    <style>
	        body { background: #18171d; font-family: sans-serif; max-width: 1300px; margin: 40px auto; padding: 0 20px; line-height: 1.6; color: #333; }
	        .item { margin-bottom: 10px; border-bottom: 1px solid #eee; padding-bottom: 10px; }
	        a { color: #0066cc; text-decoration: none; }
	        a:hover { text-decoration: underline; }
	        .date { font-size: 0.85rem; color: #666; }
	    </style>
	</head>
	<body>
	    <h1>{{.Title}}</h1>
	    <p class="date">Обновлено: {{.GeneratedAt}}</p>
	    <hr>
	    {{range $item := .Items}}
	    <div class="item">
					{{with index $item.Links 0}}
							<h4><a href="{{.Href}}" target="_blank">{{$item.Title}}</a></h4>
					{{end}}
					<!-- Если вам где-то нужна вторая ссылка (enclosure) -->
					{{if gt (len $item.Links) 1}}
							{{with index .Links 1}}
									<div class="download-zone">
											<a href="{{.Href | safeURL}}">magnet-link</a>
									</div>
							{{end}}
					{{end}}
	        <span class="date">{{.Updated}}</span>
	    </div>
	    {{end}}
	</body>
	</html>`

	// 1. Создаем FuncMap с нашей функцией
	funcMaps := template.FuncMap{
		"safeURL": func(u string) template.URL {
			return template.URL(u)
		},
	}

	// 2. Инициализируем шаблон, регистрируем функции и только ПОТОМ парсим HTML
	t, err := template.New("index").Funcs(funcMaps).Parse(tmplHTML)
	if err != nil {
		panic(err)
	}

	if err := os.MkdirAll("public", 0755); err != nil {
		panic(err)
	}

	f, err := os.Create("public/index.html")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := t.Execute(f, data); err != nil {
		panic(err)
	}

	fmt.Println("Сайт успешно сгенерирован в public/index.html")
}
