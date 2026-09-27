package tokensaver

import (
	"regexp"
	"strings"

	"github.com/dresar/go-9router/internal/storage/repos"
)

const (
	CavemanLite  = "Respond tersely. Keep grammar and full sentences but drop filler, hedging and pleasantries (just/really/basically/sure/of course/I'd be happy to). Pattern: state the thing, the action, the reason. Then next step. Code blocks, file paths, commands, errors, URLs: keep exact."
	CavemanFull  = "Respond like terse caveman. All technical substance stay exact, only fluff die. Drop articles (a/an/the), filler (just/really/basically/actually), pleasantries, hedging. Fragments OK. Pattern: [thing] [action] [reason]. [next step]. Code blocks, file paths, commands, errors, URLs: keep exact."
	CavemanUltra = "Respond ultra-terse. Maximum compression. Telegraphic. Strip conjunctions. One word when one word enough. Pattern: [thing] [action] [reason]. [next step]. Code blocks, file paths, commands, errors: keep exact."

	PonytailLite  = "You are a lazy senior developer. Lazy means efficient, not careless. The best code is the code never written. Build what's asked, but name the lazier alternative in one line. User picks."
	PonytailFull  = "You are a lazy senior developer. Lazy means efficient, not careless. The best code is the code never written. Before writing code: 1) Does this need to exist at all? (YAGNI) 2) Stdlib does it? Use it. 3) Native platform feature covers it? Use it. 4) Already-installed dependency solves it? Use it. 5) Can it be one line? One line. 6) The minimum code that works. No boilerplate. Deletion over addition. Boring over clever."
	PonytailUltra = "You are a lazy senior developer. Ultra: YAGNI extremist. Deletion before addition. Ship the one-liner and challenge the rest of the requirement in the same response. No boilerplate or scaffolding."
)

var (
	reGitDiff    = regexp.MustCompile(`(?m)^diff --git |^@@ `)
	reGitStatus  = regexp.MustCompile(`(?m)^On branch |^nothing to commit|^Changes (not |to be )|^Untracked files:`)
	reGitLog     = regexp.MustCompile(`(?m)^[*|/\\ ]*commit [0-9a-f]{7,40}`)
	reBuildOut   = regexp.MustCompile(`(?mi)^(npm (warn|error|ERR!)|yarn (warn|error)|\s*Compiling\s+\S+|\s*Downloading\s+\S+|added \d+ package|\[ERROR\]|BUILD (SUCCESS|FAILED))`)
	reTree       = regexp.MustCompile(`[├└]──|│  `)
	reLsRow      = regexp.MustCompile(`(?m)^[-dlbcps][rwx-]{9}`)
	reExcessiveWS = regexp.MustCompile(`\n{3,}`)
)

// CompressToolOutput applies RTK heuristics to compress tool results like git diff, grep, tree, build logs.
func CompressToolOutput(text string) (string, string) {
	if len(text) < 100 {
		return text, "none"
	}

	head := text
	if len(head) > 2000 {
		head = text[:2000]
	}

	filter := "generic"
	if reGitDiff.MatchString(head) {
		filter = "git-diff"
		return compressGitDiff(text), filter
	}
	if reGitStatus.MatchString(head) {
		filter = "git-status"
		return compressGitStatus(text), filter
	}
	if reGitLog.MatchString(head) {
		filter = "git-log"
		return compressGitLog(text), filter
	}
	if reBuildOut.MatchString(head) {
		filter = "build-output"
		return compressBuildOutput(text), filter
	}
	if reTree.MatchString(head) {
		filter = "tree"
		return compressTree(text), filter
	}
	if reLsRow.MatchString(head) {
		filter = "ls"
		return compressLs(text), filter
	}

	// Default deduplication / truncation for huge text (> 10000 chars)
	if len(text) > 10000 {
		lines := strings.Split(text, "\n")
		if len(lines) > 200 {
			var b strings.Builder
			for i := 0; i < 80; i++ {
				b.WriteString(lines[i])
				b.WriteString("\n")
			}
			b.WriteString(strings.Repeat("-", 40) + "\n")
			b.WriteString("... [RTK: truncated middle lines by 9router token saver] ...\n")
			b.WriteString(strings.Repeat("-", 40) + "\n")
			for i := len(lines) - 40; i < len(lines); i++ {
				b.WriteString(lines[i])
				b.WriteString("\n")
			}
			return b.String(), "smart-truncate"
		}
	}

	return reExcessiveWS.ReplaceAllString(text, "\n\n"), filter
}

func compressGitDiff(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		if strings.HasPrefix(line, "index ") || strings.HasPrefix(line, "new file mode ") || strings.HasPrefix(line, "similarity index ") {
			continue
		}
		result = append(result, line)
	}
	out := strings.Join(result, "\n")
	return reExcessiveWS.ReplaceAllString(out, "\n\n")
}

func compressGitStatus(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "(use \"git ") {
			continue
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

func compressGitLog(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		if strings.HasPrefix(line, "AuthorDate:") || strings.HasPrefix(line, "CommitDate:") {
			continue
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

func compressBuildOutput(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	skipCount := 0
	for _, line := range lines {
		if strings.Contains(line, "Downloading") || strings.Contains(line, "Fetch") || strings.Contains(line, "Extracting") {
			skipCount++
			continue
		}
		if skipCount > 0 {
			result = append(result, "... [RTK: omitted package fetch/download logs] ...")
			skipCount = 0
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

func compressTree(text string) string {
	return reExcessiveWS.ReplaceAllString(text, "\n\n")
}

func compressLs(text string) string {
	return reExcessiveWS.ReplaceAllString(text, "\n\n")
}

// ApplyTokenSaver applies Caveman, Ponytail, and RTK tool compression to the chat request body.
func ApplyTokenSaver(body map[string]any, settings repos.Settings) {
	if body == nil {
		return
	}

	rtkEnabled := repos.SettingBool(settings, "rtkEnabled", true)
	cavemanEnabled := repos.SettingBool(settings, "cavemanEnabled", false)
	cavemanLevel := repos.SettingStr(settings, "cavemanLevel", "full")
	ponytailEnabled := repos.SettingBool(settings, "ponytailEnabled", false)
	ponytailLevel := repos.SettingStr(settings, "ponytailLevel", "full")

	// 1. RTK Tool Output Compression
	if rtkEnabled {
		if msgs, ok := body["messages"].([]any); ok {
			for i, m := range msgs {
				msgMap, ok := m.(map[string]any)
				if !ok {
					continue
				}
				role, _ := msgMap["role"].(string)
				// If message is a tool output
				if role == "tool" {
					if content, ok := msgMap["content"].(string); ok && len(content) > 100 {
						compressed, _ := CompressToolOutput(content)
						msgMap["content"] = compressed
						msgs[i] = msgMap
					}
				}
			}
		}
	}

	// 2. System Prompts (Caveman & Ponytail)
	var promptAdditions []string
	if cavemanEnabled {
		switch cavemanLevel {
		case "lite":
			promptAdditions = append(promptAdditions, CavemanLite)
		case "ultra":
			promptAdditions = append(promptAdditions, CavemanUltra)
		default:
			promptAdditions = append(promptAdditions, CavemanFull)
		}
	}

	if ponytailEnabled {
		switch ponytailLevel {
		case "lite":
			promptAdditions = append(promptAdditions, PonytailLite)
		case "ultra":
			promptAdditions = append(promptAdditions, PonytailUltra)
		default:
			promptAdditions = append(promptAdditions, PonytailFull)
		}
	}

	if len(promptAdditions) == 0 {
		return
	}

	combinedSystemPrompt := strings.Join(promptAdditions, "\n\n")

	// Inject into body["messages"]
	if rawMsgs, ok := body["messages"].([]any); ok && len(rawMsgs) > 0 {
		// Look for an existing system message
		foundSystem := false
		for i, raw := range rawMsgs {
			if m, ok := raw.(map[string]any); ok {
				if r, _ := m["role"].(string); r == "system" {
					if content, ok := m["content"].(string); ok {
						m["content"] = content + "\n\n" + combinedSystemPrompt
					} else {
						m["content"] = combinedSystemPrompt
					}
					rawMsgs[i] = m
					foundSystem = true
					break
				}
			}
		}
		if !foundSystem {
			// Prepend new system message
			newSystem := map[string]any{
				"role":    "system",
				"content": combinedSystemPrompt,
			}
			body["messages"] = append([]any{newSystem}, rawMsgs...)
		}
	}
}
