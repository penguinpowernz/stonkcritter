package renderers

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/penguinpowernz/stonkcritter/models"
)

func Mastodon(w io.Writer, dd models.Disclosures) error {
	fmt.Fprintln(w, mastodonCritters(dd.Critters()))
	// TODO: reverse the order
	dd.SortBy(new(models.SortByDisclosureDate), false).Each(func(d models.Disclosure) {
		fmt.Fprintln(w, mastodonPost()(d))
	})
	return nil
}

func MastodonText() func(d models.Disclosure) string {
	return func(d models.Disclosure) string {
		desc := "#" + d.TickerString()
		assetdesc := " (" + d.AssetDescription + ")"
		if desc == "#??" {
			desc = d.AssetDescription
			assetdesc = ""
		}

		assettyp := d.AssetType
		if assettyp == "" {
			assettyp = "asset"
		}

		if assettyp == "Other Securities" && strings.Contains(assetdesc, "ETF") {
			assettyp = "ETF"
		}

		if assettyp == "Other Securities" {
			assettyp = "security"
		}

		verb := ""
		switch d.TradeType() {
		case "sale (partial)", "sale_partial":
			verb = "partially sold"
		case "sale", "sale_full":
			verb = "sold"
		case "purchase":
			verb = "purchased"
		case "exchange":
			verb = "exchanged"
		default:
			verb = d.TradeType()
		}

		assetdesc = strings.ReplaceAll(assetdesc, " Common Stock", "")
		desc = strings.ReplaceAll(desc, " Common Stock", "")

		daysago := fmt.Sprintf("%d", int((d.DisclosedOn().Unix()-d.TransactionOn().Unix())/86400))
		switch daysago {
		case "0":
			daysago = "Today"
		case "1":
			daysago = "Yesterday"
		default:
			daysago = daysago + " days ago"
		}

		who := "I"
		switch strings.ToLower(d.Owner) {
		case "child", "dependent":
			who = "my child"
		case "spouse":
			who = "my spouse"
		}

		var text string
		if assettyp == "PDF Disclosed Filing" {
			text = fmt.Sprintf("%s %s made a trade.  I have disclosed this in a scanned PDF at %s", daysago, who, d.PtrLink)
		} else {
			text = fmt.Sprintf("%s %s %s the %s; %s for %s%s %s", daysago, who, verb, assettyp, desc, d.Amount, assetdesc, d.PtrLink)
		}

		p := bluemonday.StripTagsPolicy()
		text = p.Sanitize(text)
		text = strings.ReplaceAll(text, "&amp;", "&")
		text = strings.ReplaceAll(text, "\"", "'")
		return text
	}
}

func mastodonPost() func(d models.Disclosure) string {
	data := `PostStatusService.new.call({{.CritterP}}, text: "{{.Text}}").tap {|s| s.created_at = DateTime.parse("{{.Date}}") }.save`

	tmpl, err := template.New("").Parse(data)
	if err != nil {
		panic(err)
	}

	render := MastodonText()

	return func(d models.Disclosure) string {
		var title string
		if d.Representative != "" {
			title = "Rep."
		}

		if d.Senator != "" {
			title = "Sen."

		}
		view := struct {
			Text     string
			Critter  string
			CritterP string
			Date     string
		}{render(d), title + " " + d.CritterName(), parameterize(d.CritterName()), d.TransactionOn().Format(time.RFC3339)}
		b := strings.Builder{}
		if err := tmpl.Execute(&b, &view); err != nil {
			panic(err)
		}
		return b.String()
	}
}

func parameterize(x string) string {
	x = strings.ToLower(x)
	re := regexp.MustCompile(`[^\w]`)
	re2 := regexp.MustCompile(`_+`)
	x = re.ReplaceAllString(x, "_")
	x = re2.ReplaceAllString(x, "_")
	return x
}

func mastodonCritters(critters []string) string {
	var lines []string

	data := `{{.Username}} = Account.find_or_create_by(username: "{{.Username}}", display_name: "{{.Name}}", discoverable: true, avatar_filename: "{{.Avatar}}", avatar_file_size: 0, avatar_updated_at: DateTime.now)
{{.Username}}.user = User.create(email: "{{.Username}}@example.com", sign_up_ip: "127.0.0.1", password: "{{.Password}}", password_confirmation: "{{.Password}}", confirmed_at: DateTime.now, agreement: true) if {{.Username}}.user.nil?
{{.Username}}.save`

	tmpl := template.New("account")
	tmpl.Parse(data)

	for _, name := range critters {
		name = strings.ReplaceAll(name, "\"", "'")
		b := make([]byte, 32)
		rand.Read(b)
		pw := fmt.Sprintf("%x", string(b))
		uname := parameterize(name)

		re := regexp.MustCompile(`^[^\.]*\. `)
		bits := strings.Split(re.ReplaceAllString(name, ""), " ")
		fname := bits[0]
		lname := bits[len(bits)-1]
		avatar := strings.ToLower(fname + "_" + lname)

		view := struct {
			Name     string
			Username string
			Password string
			Avatar   string
		}{name, uname, pw, avatar}

		buf := strings.Builder{}
		tmpl.Execute(&buf, view)
		lines = append(lines, buf.String())
	}

	return strings.Join(lines, "\n")
}
