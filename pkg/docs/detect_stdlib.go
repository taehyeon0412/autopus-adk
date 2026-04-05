package docs

// goStdLib is the set of known Go standard library module names and paths.
var goStdLib = map[string]bool{
	"fmt": true, "os": true, "io": true, "net": true, "net/http": true,
	"path": true, "path/filepath": true, "strings": true, "strconv": true,
	"errors": true, "context": true, "sync": true, "time": true,
	"encoding": true, "encoding/json": true, "encoding/xml": true,
	"crypto": true, "log": true, "sort": true, "bytes": true,
	"bufio": true, "regexp": true, "reflect": true, "testing": true,
	"flag": true, "math": true, "unicode": true, "html": true,
	"archive": true, "compress": true, "database": true, "debug": true,
	"embed": true, "go": true, "hash": true, "image": true,
	"index": true, "mime": true, "plugin": true, "runtime": true,
	"syscall": true, "text": true, "unsafe": true,
}

// nodeStdLib is the set of known Node.js built-in module names.
var nodeStdLib = map[string]bool{
	"fs": true, "path": true, "http": true, "https": true, "os": true,
	"url": true, "util": true, "crypto": true, "stream": true,
	"events": true, "child_process": true, "buffer": true, "assert": true,
	"cluster": true, "dgram": true, "dns": true, "net": true,
	"readline": true, "repl": true, "tls": true, "vm": true,
	"zlib": true, "console": true, "module": true, "process": true,
	"querystring": true, "string_decoder": true, "timers": true,
	"tty": true, "v8": true, "worker_threads": true,
}

// pythonStdLib is the set of known Python standard library module names.
var pythonStdLib = map[string]bool{
	"os": true, "sys": true, "json": true, "re": true, "math": true,
	"datetime": true, "collections": true, "itertools": true,
	"functools": true, "typing": true, "pathlib": true, "io": true,
	"shutil": true, "subprocess": true, "argparse": true, "logging": true,
	"unittest": true, "abc": true, "copy": true, "enum": true,
	"dataclasses": true, "hashlib": true, "hmac": true, "secrets": true,
	"socket": true, "http": true, "urllib": true, "email": true,
	"html": true, "xml": true, "csv": true, "sqlite3": true,
	"pickle": true, "struct": true, "threading": true,
	"multiprocessing": true, "asyncio": true, "concurrent": true,
	"ctypes": true, "inspect": true, "importlib": true, "pkgutil": true,
	"pdb": true, "traceback": true, "warnings": true, "contextlib": true,
	"textwrap": true, "string": true, "random": true, "statistics": true,
	"decimal": true, "fractions": true, "operator": true, "array": true,
	"bisect": true, "heapq": true, "queue": true, "weakref": true,
	"types": true, "dis": true, "ast": true, "compileall": true,
	"py_compile": true, "pprint": true, "tempfile": true, "glob": true,
	"fnmatch": true, "linecache": true, "tokenize": true, "difflib": true,
	"base64": true, "binascii": true, "codecs": true, "unicodedata": true,
	"locale": true, "gettext": true, "calendar": true, "time": true,
	"sched": true, "signal": true, "mmap": true, "select": true,
	"selectors": true, "platform": true, "errno": true, "faulthandler": true,
	"gc": true, "site": true, "builtins": true,
}

// goKnownLibs is the set of popular Go third-party library names for text detection.
var goKnownLibs = map[string]bool{
	"cobra": true, "viper": true, "gin": true, "fiber": true, "echo": true,
	"chi": true, "mux": true, "gorm": true, "sqlx": true, "zap": true,
	"zerolog": true, "logrus": true, "testify": true, "mock": true,
	"wire": true, "fx": true, "prometheus": true, "grpc": true,
	"protobuf": true, "uuid": true, "jwt": true, "bcrypt": true,
	"validator": true, "pflag": true, "afero": true, "cast": true,
}
