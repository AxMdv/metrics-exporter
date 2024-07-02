package Data

type Matching struct {
	LabelId string
	Plu     []string
}

type StatusMessage struct {
	BOStatus  bool
	ESLStatus bool
}

type Config struct {
	BoIp           string
	BoPort         string
	Workers        int
	CrawlerTimeout int
	StatusTimeout  int
	DumpFlag       string
	Debug          bool
	Token          string
}
