package plugins

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/johnwilson/restapi/system"
)

type PluginQM struct {
	qm *QueryManager
}

func (qp *PluginQM) Init(a *system.Application) error {

	qm := newQueryManager()

	// get config
	fp := a.Config.Get("sqlqueries.path").(string)
	f, err := os.Open(fp)
	defer f.Close()

	if err != nil {
		return fmt.Errorf("query manager: sql file loading failed:\n%s", err)
	}

	// load queries
	qm.Load(f)
	qp.qm = qm

	return nil
}

func (qp *PluginQM) Get() interface{} {
	return qp.qm
}

func (qp *PluginQM) Close() error {
	return nil
}

type QueryManager struct {
	queries map[string]string
}

func newQueryManager() *QueryManager {
	qm := new(QueryManager)
	qm.queries = map[string]string{}
	return qm
}

func (qm *QueryManager) Load(r io.Reader) {
	scanner := SQLScanner{}
	q := scanner.Run(bufio.NewScanner(r))

	for k, v := range q {
		qm.queries[k] = v
	}
}

func (qm *QueryManager) Get(name string) string {
	q, ok := qm.queries[name]
	if !ok {
		return ""
	}
	return q
}

type SQLScanner struct {
	line    string
	queries map[string]string
	current string
}

type stateFn func(*SQLScanner) stateFn

func getTag(line string) string {
	re := regexp.MustCompile("^\\s*--\\s*name:\\s*(\\S+)")
	matches := re.FindStringSubmatch(line)
	if matches == nil {
		return ""
	}
	return matches[1]
}

func initialState(s *SQLScanner) stateFn {
	if tag := getTag(s.line); len(tag) > 0 {
		s.current = tag
		return queryState
	}
	return initialState
}

func queryState(s *SQLScanner) stateFn {
	if tag := getTag(s.line); len(tag) > 0 {
		s.current = tag
	} else {
		s.appendQueryLine()
	}
	return queryState
}

func (s *SQLScanner) appendQueryLine() {
	current := s.queries[s.current]
	line := strings.Trim(s.line, " \t")
	if len(line) == 0 {
		return
	}

	if len(current) > 0 {
		current = current + "\n"
	}

	current = current + line
	s.queries[s.current] = current
}

func (s *SQLScanner) Run(io *bufio.Scanner) map[string]string {
	s.queries = make(map[string]string)

	for state := initialState; io.Scan(); {
		s.line = io.Text()
		state = state(s)
	}

	return s.queries
}
