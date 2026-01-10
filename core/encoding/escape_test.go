package encoding

import "testing"

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"plain text", "Hello World", "Hello World"},
		{"ampersand", "Tom & Jerry", "Tom &amp; Jerry"},
		{"less than", "a < b", "a &lt; b"},
		{"greater than", "a > b", "a &gt; b"},
		{"quotes", `He said "hello"`, "He said &#34;hello&#34;"},
		{"apostrophe", "it's", "it&#39;s"},
		{"multiple", `<tag attr="value">content & more</tag>`, "&lt;tag attr=&#34;value&#34;&gt;content &amp; more&lt;/tag&gt;"},
		{"unicode", "日本語 & émoji 🎉", "日本語 &amp; émoji 🎉"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeXML(tt.input)
			if got != tt.want {
				t.Errorf("EscapeXML(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEscapeXMLText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"plain text", "Hello World", "Hello World"},
		{"ampersand", "Tom & Jerry", "Tom &amp; Jerry"},
		{"less than", "a < b", "a &lt; b"},
		{"greater than", "a > b", "a &gt; b"},
		{"quotes preserved", `He said "hello"`, `He said "hello"`},
		{"all three", "<script>&</script>", "&lt;script&gt;&amp;&lt;/script&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeXMLText(tt.input)
			if got != tt.want {
				t.Errorf("EscapeXMLText(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEscapeXMLAttr(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"plain text", "Hello World", "Hello World"},
		{"ampersand", "Tom & Jerry", "Tom &amp; Jerry"},
		{"double quotes", `He said "hello"`, "He said &quot;hello&quot;"},
		{"all chars", `<tag attr="val&ue">`, "&lt;tag attr=&quot;val&amp;ue&quot;&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeXMLAttr(tt.input)
			if got != tt.want {
				t.Errorf("EscapeXMLAttr(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"plain text", "Hello World", "Hello World"},
		{"ampersand", "Tom & Jerry", "Tom &amp; Jerry"},
		{"less than", "a < b", "a &lt; b"},
		{"greater than", "a > b", "a &gt; b"},
		{"quotes", `He said "hello"`, "He said &quot;hello&quot;"},
		{"html tag", "<script>alert('xss')</script>", "&lt;script&gt;alert('xss')&lt;/script&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeHTML(tt.input)
			if got != tt.want {
				t.Errorf("EscapeHTML(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEscapeLaTeX(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"plain text", "Hello World", "Hello World"},
		{"backslash", `path\to\file`, `path\textbackslash{}to\textbackslash{}file`},
		{"braces", "{curly}", `\{curly\}`},
		{"dollar", "price $100", `price \$100`},
		{"percent", "100% complete", `100\% complete`},
		{"ampersand", "Tom & Jerry", `Tom \& Jerry`},
		{"hash", "#1 best", `\#1 best`},
		{"underscore", "var_name", `var\_name`},
		{"caret", "x^2", `x\^{}2`},
		{"tilde", "~user", `\~{}user`},
		{"multiple", `$100 & {test}`, `\$100 \& \{test\}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeLaTeX(tt.input)
			if got != tt.want {
				t.Errorf("EscapeLaTeX(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEscapeManifest(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"plain text", "Hello World", "Hello World"},
		{"newline", "line1\nline2", "line1 line2"},
		{"carriage return", "line1\rline2", "line1 line2"},
		{"crlf", "line1\r\nline2", "line1  line2"},
		{"multiple newlines", "a\nb\nc\n", "a b c "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeManifest(tt.input)
			if got != tt.want {
				t.Errorf("EscapeManifest(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
