package snapshots

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
	"strings"
	"strconv"
	"path/filepath"
	"encoding/json"
	"github.com/bmatcuk/doublestar/v4"
)

type Snapshot struct {
	Message       string
	Time          string
	Files	      []string
	StoredFiles	  map[string]string
	ModTimes	  map[string]string
}


