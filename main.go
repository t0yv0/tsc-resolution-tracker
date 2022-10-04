package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type tsImport struct {
	fromFile     string
	modName      string
	resolvedFile string
	resolved     bool
}

func main() {
	why := flag.String("why", "", "regex pattern for a module name to explain why it was imported")
	flag.Parse()

	if why != nil && *why != "" {
		if err := runTypeScriptAndPrintWhyModuleImported(*why); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := runTypescriptAndPrintParsedImpors(); err != nil {
		log.Fatal(err)
	}
}

func prepareModulePathNormalizer() func(string) string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	prefixToTrim := []string{
		strings.TrimSuffix(filepath.Join(wd, "node_modules", "x"), "x"),
		strings.TrimSuffix(filepath.Join(wd, "x"), "x"),
	}
	trim := func(x string) string {
		for _, p := range prefixToTrim {
			x = strings.TrimPrefix(x, p)
		}
		return x
	}
	return trim
}

func runTypeScriptAndPrintWhyModuleImported(whyPattern string) error {
	var matches func(modName string) bool

	pat, err := regexp.Compile(whyPattern)
	if err != nil {
		matches = func(modName string) bool {
			return modName == whyPattern
		}
	}

	matches = func(modName string) bool {
		return len(pat.FindStringIndex(modName)) > 0
	}

	trim := prepareModulePathNormalizer()

	graph := &Graph{}

	var root string

	parsed := func(i tsImport) error {
		if i.resolved {
			from := trim(i.fromFile)
			to := trim(i.resolvedFile)
			if root == "" {
				root = from
			}
			graph.addEdge(from, to)
		}
		return nil
	}

	if err := runTypescriptAndParseImports(parsed); err != nil {
		return err
	}

	rootNode := graph.node(root)

	matched := 0

	for label, node := range graph.nodes {
		if matches(label) {
			matched++
			p := graph.findPath(rootNode, node)
			if p != nil {
				fmt.Printf("%s\n\n", strings.Join(p.slice(), " ->\n"))
			}
		}
	}

	fmt.Printf("total modules matching %q: %d\n", whyPattern, matched)
	return nil
}

func runTypescriptAndPrintParsedImpors() error {
	trim := prepareModulePathNormalizer()

	parsed := func(i tsImport) error {
		if i.resolved {
			fmt.Printf("%s -> %s\n",
				trim(i.fromFile),
				trim(i.resolvedFile))
		}
		return nil
	}

	return runTypescriptAndParseImports(parsed)
}

func runTypescriptAndParseImports(parse func(tsImport) error) error {
	tsc, err := exec.LookPath("tsc")
	if err != nil {
		return err
	}
	cmd := exec.Command(tsc, "--traceResolution")

	out, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("cmd.StodutPipe() failed: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cmd.Start() failed: %w", err)
	}

	if err := parseImports(out, parse); err != nil {
		return fmt.Errorf("parseImports() failed: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("cmd.Wait() failed: %w", err)
	}

	return nil
}

func parseImports(reader io.Reader, parse func(tsImport) error) error {
	queries := map[string]tsImport{}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "====") {
			if ok, r := matchResolving(line); ok {
				if _, exist := queries[r.modName]; exist {
					return fmt.Errorf("Already seen %v", r.modName)
				}
				queries[r.modName] = r
			} else if ok, r := matchResolved(line); ok {
				q, got := queries[r.modName]
				if !got {
					return fmt.Errorf("Resolved but not queried %v", r.modName)
				}
				q.resolved = true
				q.resolvedFile = r.resolvedFile
				delete(queries, r.modName)
				if err := parse(q); err != nil {
					return err
				}
			} else if ok, m := matchNotResolved(line); ok {
				q, got := queries[m]
				delete(queries, m)
				if got {
					if err := parse(q); err != nil {
						return err
					}
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("parseImports failed to scan text: %w", err)
	}
	return nil
}

var resolvingRegex *regexp.Regexp = regexp.MustCompile(`======== Resolving module '(?P<ImportSpec>[^']+)' from '(?P<FromFile>[^']+)'. ========`)

func matchResolving(line string) (bool, tsImport) {
	match := resolvingRegex.FindStringSubmatch(line)
	if match == nil {
		return false, tsImport{}
	}
	var importSpec, fromFile string
	for i, n := range resolvingRegex.SubexpNames() {
		if n == "ImportSpec" {
			importSpec = match[i]
		}
		if n == "FromFile" {
			fromFile = match[i]
		}
	}

	return true, tsImport{
		fromFile: fromFile,
		modName:  importSpec,
	}
}

var resolvedRegex *regexp.Regexp = regexp.MustCompile(`======== Module name '(?P<ModName>[^']+)' was successfully resolved to '(?P<ResolvedFile>[^']+)'`)

func matchResolved(line string) (bool, tsImport) {
	match := resolvedRegex.FindStringSubmatch(line)
	if match == nil {
		return false, tsImport{}
	}
	var modName, resolvedFile string
	for i, n := range resolvedRegex.SubexpNames() {
		if n == "ModName" {
			modName = match[i]
		}
		if n == "ResolvedFile" {
			resolvedFile = match[i]
		}
	}
	return true, tsImport{
		modName:      modName,
		resolvedFile: resolvedFile,
	}
}

var notResolvedRegex *regexp.Regexp = regexp.MustCompile(`======== Module name '(?P<ModName>[^']+)' was not resolved. ========`)

func matchNotResolved(line string) (bool, string) {
	match := notResolvedRegex.FindStringSubmatch(line)
	if match == nil {
		return false, ""
	}
	var modName string
	for i, n := range notResolvedRegex.SubexpNames() {
		if n == "ModName" {
			modName = match[i]
		}
	}
	return true, modName
}

func noError(err error) {
	if err != nil {
		panic(err)
	}
}
