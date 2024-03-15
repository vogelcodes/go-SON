package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
    templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
    return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
    return &Templates{
        templates: template.Must(template.ParseGlob("views/*.html")),
    }
}

type Contact struct {
    Name string
    Email string
}

func newContact(name, email string) Contact {
    return Contact{
        Name: name,
        Email: email,
    }
}

type Contacts = []Contact

type Data struct {
    Contacts Contacts
}

func (d *Data) hasEmail(email string) bool {
    for _, contact := range d.Contacts {
        if contact.Email == email {
            return true
        }
    }
    return false
}

func newData() Data {
    return Data{
        Contacts: []Contact{
            newContact("John", "aoeu"),
            newContact("Clara", "cd@gmail.com"),
        },
    }
}

type FormData struct {
    Values map[string]string
    Errors map[string]string
}

func newFormData() FormData {
    return FormData{
        Values: make(map[string]string),
        Errors: make(map[string]string),
    }
}

type Page struct {
    Data Data
    Form FormData
}

type Lead struct {
    Email string
    Phone string
    Name string
    Tag string
    Date string
    Cta string
    Url string
    AvatarUrl string
    Location string
}

func newPage() Page {
    return Page{
        Data: newData(),
        Form: newFormData(),
    }
}

func main() {

    e := echo.New()
    e.Use(middleware.Logger())
    e.Static("/images", "images")
    e.Static("/css", "css")
    e.Static("/fonts", "fonts")
    e.Static("/js", "js")

    page := newPage()
    e.Renderer = newTemplate()

    e.GET("/", func(c echo.Context) error {
        return c.Render(200, "index", page)
    })
    e.GET("/leads", func(c echo.Context) error {
        response, err := http.Get("https://script.google.com/macros/s/AKfycbwIAj1HWYmqEeF7I_A3WfJGoshnPzSbQLDYir00RhgoWs1QsRj5nLAsEUIAGYuD7DfopQ/exec")
        if err != nil {
            fmt.Println(err)
            return c.String(500, "Erro ao buscar leads") // HTTP Status 500 para erro 
        }
        defer response.Body.Close() // Garante o fechamento do Body
    
        bodyBytes, err := io.ReadAll(response.Body)
        var leads [][]interface{}   
        json.Unmarshal(bodyBytes, &leads)
        //remove the first element
        leads = leads[1:]
        //create a new array with struct Lead iterating over the leads
        var leadsStruct []Lead
        for _, lead := range leads {
            var cta string
            switch v := lead[5].(type) {
            case float64:
                cta = strconv.FormatFloat(v, 'f', -1, 64)
            case string:
                cta = v
            default:
                // handle error
            }
            
            leadsStruct = append(leadsStruct, Lead{
                Email: lead[0].(string),
                Phone: lead[1].(string),
                Name: lead[2].(string),
                Tag: lead[3].(string),
                Date: lead[4].(string),
                Cta: cta,
                Url: lead[6].(string),
                AvatarUrl: lead[7].(string),
                Location: lead[8].(string),
            })
        }
        
        return c.JSON(200, leadsStruct[0])
    })

    e.POST("/contacts", func(c echo.Context) error {
        name := c.FormValue("name")
        email := c.FormValue("email")

        if page.Data.hasEmail(email) {
            formData := newFormData()
            formData.Values["name"] = name
            formData.Values["email"] = email
            formData.Errors["email"] = "Email already exists"

            return c.Render(422, "form", formData)
        }

        contact := newContact(name, email)
        page.Data.Contacts = append(page.Data.Contacts, contact)

        // TODO: ??????
        c.Render(200, "form", newFormData())
        return c.Render(200, "oob-contact", contact)
    })


    e.Logger.Fatal(e.Start(":42069"))

}
