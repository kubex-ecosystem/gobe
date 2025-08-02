package types

type DataImporter interface {
	ImportFromYAML(filename string) error
	ImportFromJSON(filename string) error
	ImportFromXML(filename string) error
	ImportFromTOML(filename string) error
	ImportFromENV(filename string) error
	ImportFromINI(filename string) error
	ImportFromCSV(filename string) error
	ImportFromProperties(filename string) error
	ImportFromText(filename string) error
	ImportFromASN(filename string) error
	ImportFromBinary(filename string) error
	ImportFromHTML(filename string) error
	ImportFromExcel(filename string) error
	ImportFromPDF(filename string) error
	ImportFromMarkdown(filename string) error
}
type dataImporter struct{}

func NewDataImporter() DataImporter { return &dataImporter{} }

func (d dataImporter) ImportFromYAML(filename string) error {
	return nil
}
func (d dataImporter) ImportFromJSON(filename string) error {
	return nil
}
func (d dataImporter) ImportFromXML(filename string) error {
	return nil
}
func (d dataImporter) ImportFromTOML(filename string) error {
	return nil
}
func (d dataImporter) ImportFromENV(filename string) error {
	return nil
}
func (d dataImporter) ImportFromINI(filename string) error {
	return nil
}

func (d dataImporter) ImportFromCSV(filename string) error {
	return nil
}
func (d dataImporter) ImportFromProperties(filename string) error {
	return nil
}
func (d dataImporter) ImportFromText(filename string) error {
	return nil
}
func (d dataImporter) ImportFromASN(filename string) error {
	return nil
}
func (d dataImporter) ImportFromHTML(filename string) error {
	return nil
}
func (d dataImporter) ImportFromMarkdown(filename string) error {
	return nil
}

func (d dataImporter) ImportFromBinary(filename string) error {
	return nil
}
func (d dataImporter) ImportFromExcel(filename string) error {
	return nil
}
func (d dataImporter) ImportFromPDF(filename string) error {
	return nil
}
