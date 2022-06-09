package perf

type Table struct {
	title        []string
	rows         [][]string
	alignedTitle []string
	alignedRows  [][]string
}

func (t *Table) SetTitle(title []string) {
	t.title = title
}

func (t *Table) AddRow(row []string) {
	t.rows = append(t.rows, row)
}

func (t *Table) String() string {
	s := ""

	maxLen := []int{}
	columnNum := len(t.title)

	for _, v := range t.title {
		maxLen = append(maxLen, len(v))
	}

	for _, v := range t.rows {
		for j, vv := range v {
			if len(maxLen) < j+1 {
				maxLen = append(maxLen, len(vv))
			} else if len(vv) > maxLen[j] {
				maxLen[j] = len(vv)
			}
		}
		if len(v) > columnNum {
			columnNum = len(v)
		}
	}

	for len(t.title) < columnNum {
		t.title = append(t.title, "")
	}

	for i, v := range t.rows {
		for len(v) < columnNum {
			t.rows[i] = append(t.rows[i], "")
		}
	}

	for i, _ := range maxLen {
		maxLen[i] += 2
	}

	s += "|"
	t.alignedTitle = make([]string, len(t.title))
	for i, v := range t.title {
		t.alignedTitle[i] = padding(v, maxLen[i])
		s += t.alignedTitle[i]
		s += "|"
	}
	s += "\n"

	t.alignedRows = make([][]string, len(t.rows))
	for i, v := range t.rows {
		rows := make([]string, len(v))
		s += "|"
		for j, vv := range v {
			rows[j] = padding(vv, maxLen[j])
			s += rows[j]
			s += "|"
		}
		s += "\n"
		t.alignedRows[i] = rows
	}

	return s
}

func padding(s string, maxLen int) string {
	paddingLen := maxLen - len(s)
	if paddingLen > 0 {
		for i := 0; i < paddingLen/2; i++ {
			s = " " + s
		}
		for i := 0; i < paddingLen/2; i++ {
			s += " "
		}
		if paddingLen%2 == 1 {
			s += " "
		}
	}
	return s
}

func NewTable() *Table {
	return &Table{}
}
