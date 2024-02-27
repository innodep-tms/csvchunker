package csvutil

import (
	"encoding/csv"
	"errors"
	"net/http"
	"reflect"
	"strconv"

	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Chunker[T any] struct {
	GinCtx *gin.Context
	Writer *csv.Writer
}

func NewChunker[T any](ginContext *gin.Context) *Chunker[T] {
	return &Chunker[T]{
		GinCtx: ginContext,
		Writer: csv.NewWriter(ginContext.Writer),
	}
}

// SetHeader sets the response header for the CSV file.
// The filename is the name of the file to be downloaded.
// The default encoding is UTF-8.
// Transfer-Encoding is set to chunked.
func (chunker *Chunker[T]) SetHeader(filename string) {
	// Set response header
	chunker.GinCtx.Status(http.StatusOK)
	chunker.GinCtx.Header("Content-Type", "text/csv; charset=utf-8")
	chunker.GinCtx.Header("Transfer-Encoding", "chunked")
	chunker.GinCtx.Header("Content-Disposition", `attachment;filename="`+filename+`.csv"`)
	chunker.GinCtx.Header("Content-Description", "File Transfer")

	// TODO: make it possible to select BOM
	// Now, use BOM of UTF-8 representation
	chunker.GinCtx.Writer.Write([]byte("\xEF\xBB\xBF"))
}

// WriteCsvLabel writes the Label of the CSV data.
func (chunker *Chunker[T]) WriteCsvLabel() {
	t := new(T)
	labels := []string{}
	for _, f := range structs.Fields(t) {
		if label := f.Tag("csv"); label != "" {
			labels = append(labels, label)
		}
	}

	chunker.Writer.Write(labels)
}

// WriteChunk writes the data in chunks.
func (chunker *Chunker[T]) WriteChunk(data []T) error {
	for i := range data {
		row := []string{}
		for _, f := range structs.Fields(data[i]) {
			if label := f.Tag("csv"); label != "" {
				value := ""
				switch reflect.TypeOf(f.Value()).String() {
				case "string":
					value = f.Value().(string)
				case "*string":
					value = *f.Value().(*string)
				case "int":
					value = strconv.Itoa(f.Value().(int))
				case "*int":
					value = strconv.Itoa(*f.Value().(*int))
				case "int64":
					value = strconv.FormatInt(f.Value().(int64), 10)
				case "*int64":
					value = strconv.FormatInt(*f.Value().(*int64), 10)
				case "float32":
					value = strconv.FormatFloat(float64(f.Value().(float32)), 'f', -6, 32)
				case "*float32":
					value = strconv.FormatFloat(float64(*f.Value().(*float32)), 'f', -6, 32)
				case "float64":
					value = strconv.FormatFloat(f.Value().(float64), 'f', -6, 64)
				case "*float64":
					value = strconv.FormatFloat(*f.Value().(*float64), 'f', -6, 64)
				}

				row = append(row, value)
			}
		}

		chunker.Writer.Write(row)
	}

	return nil
}

// ResetWriter resets the csv writer.
func (chunker *Chunker[T]) ResetWriter() {
	chunker.Writer.Flush()
	chunker.Writer = csv.NewWriter(chunker.GinCtx.Writer)
}

// TransferChunk is a function that works as follows:
// 1. Fetches the data from the cursor.
// 2. Writes the data in chunks.
// 3. Resets the writer.
// Need to Chunker and Cursor to use this function.
func TransferChunk[T any](chunker *Chunker[T], cursor *Cursor[T]) error {
	if chunker == nil || chunker.GinCtx == nil || cursor == nil || cursor.DBconn == nil {
		return errors.New("Need to initialize chunker and cursor")
	}

	for data, err := cursor.FetchCursor(); err == nil && len(data) > 0; data, err = cursor.FetchCursor() {
		if err = chunker.WriteChunk(data); err != nil {
			return err
		}
		chunker.ResetWriter()
	}

	return nil
}

// TransferCSVFileChunked is a function that responses a large csv file .
// The function works as follows:
// 1. Initializes the chunker and cursor.
// 2. Writes the header and the label of the CSV file.
// 3. Fetches the data from the cursor and writes the data in chunks.
// 4. Resets the writer.
// Need to gin.Context and gorm.DB to use this function.
// The fetchSize is the number of rows to be fetched at a time.
func TransferCSVFileChunked[T any](
	ginContext *gin.Context, dbconn *gorm.DB,
	query, filename string,
	chunkSize int,
) error {

	chunker := NewChunker[T](ginContext)
	defer chunker.Writer.Flush()

	cursor, err := NewCursor[T](dbconn, query, chunkSize)
	if err != nil {
		return err
	}
	defer cursor.Close()

	chunker.SetHeader(filename)
	chunker.WriteCsvLabel()

	if err := TransferChunk[T](chunker, cursor); err != nil {
		return err
	}

	return nil
}
