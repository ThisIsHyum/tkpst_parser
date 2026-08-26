package replaces

import (
	"errors"
	"io"
	"os"

	docx "github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/domain"
)

func getTableFromWord(file io.ReadCloser) (domain.Table, error) {
	if file == nil {
		return nil, errors.New("file is nil")
	}
	defer file.Close()
	d, err := docx.OpenDocumentFromReader(file)
	if err != nil {
		return nil, err
	}
	tables := d.Tables()
	if len(tables) == 0 {
		return nil, errors.New("tables not exists")
	}
	return tables[0], nil
}

func GetTablefromFile(path string) (domain.Table, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return getTableFromWord(file)
}
