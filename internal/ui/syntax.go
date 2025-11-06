package ui

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// detectLanguage detects the programming language from file extension
func detectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Map common extensions to Chroma language names
	languageMap := map[string]string{
		".go":     "go",
		".js":     "javascript",
		".ts":     "typescript",
		".jsx":    "jsx",
		".tsx":    "tsx",
		".py":     "python",
		".java":   "java",
		".c":      "c",
		".cpp":    "cpp",
		".cc":     "cpp",
		".cxx":    "cpp",
		".h":      "c",
		".hpp":    "cpp",
		".cs":     "csharp",
		".rb":     "ruby",
		".php":    "php",
		".rs":     "rust",
		".swift":  "swift",
		".kt":     "kotlin",
		".scala":  "scala",
		".sh":     "bash",
		".bash":   "bash",
		".zsh":    "bash",
		".ps1":    "powershell",
		".yaml":   "yaml",
		".yml":    "yaml",
		".json":   "json",
		".xml":    "xml",
		".html":   "html",
		".css":    "css",
		".scss":   "scss",
		".sass":   "sass",
		".sql":    "sql",
		".md":     "markdown",
		".vim":    "vim",
		".lua":    "lua",
		".pl":     "perl",
		".r":      "r",
		".m":      "objective-c",
		".dart":   "dart",
		".ex":     "elixir",
		".exs":    "elixir",
		".erl":    "erlang",
		".hs":     "haskell",
		".clj":    "clojure",
		".ml":     "ocaml",
		".fs":     "fsharp",
		".vb":     "vb.net",
		".groovy": "groovy",
		".gradle": "groovy",
	}

	if lang, ok := languageMap[ext]; ok {
		return lang
	}

	// Try to detect by filename
	filename := strings.ToLower(filepath.Base(filePath))
	if filename == "makefile" || filename == "gnumakefile" {
		return "make"
	}
	if filename == "dockerfile" || strings.HasPrefix(filename, "dockerfile.") {
		return "docker"
	}
	if filename == "jenkinsfile" {
		return "groovy"
	}

	return ""
}

// highlightCode applies syntax highlighting to a single line of code
func highlightCode(code, language string) string {
	if language == "" || code == "" {
		return code
	}

	// Get lexer for the language
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// Use a terminal-friendly style (monokai is good for dark terminals)
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	// Create a terminal formatter with 256 colors
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	// Tokenize and format
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return code
	}

	// Return the highlighted code, trimming trailing newline
	result := buf.String()
	return strings.TrimSuffix(result, "\n")
}

// highlightDiffLine applies syntax highlighting to a diff line
// It preserves the diff prefix (+, -, space) and applies highlighting to the code part
func highlightDiffLine(line, language string) string {
	if len(line) == 0 {
		return line
	}

	// Extract the diff prefix (first character)
	prefix := ""
	codePart := line

	if len(line) > 0 {
		firstChar := line[0]
		if firstChar == '+' || firstChar == '-' || firstChar == ' ' {
			prefix = string(firstChar)
			if len(line) > 1 {
				codePart = line[1:]
			} else {
				codePart = ""
			}
		}
	}

	// Apply syntax highlighting to the code part
	if codePart != "" && language != "" {
		codePart = highlightCode(codePart, language)
	}

	return prefix + codePart
}
