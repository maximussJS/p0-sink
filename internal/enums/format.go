package enums

type DatasetFormat string

const (
	DatasetFormatJSON    DatasetFormat = "json"
	DatasetFormatCSV     DatasetFormat = "csv"
	DatasetFormatParquet DatasetFormat = "parquet"
)
