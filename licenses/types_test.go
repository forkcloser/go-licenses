package licenses

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/forkcloser/go-licenses/internal/third_party/licenseclassifier/assets"
)

func TestTypes(t *testing.T) {
	classifier, err := assets.DefaultClassifier()
	if err != nil {
		t.Fatal(err)
	}

	fv := reflect.ValueOf(classifier).Elem().FieldByName("docs")

	uniqueNames := make(map[string]struct{})

	for _, key := range fv.MapKeys() {
		keyValue := key.String()

		splits := strings.Split(keyValue, string(os.PathSeparator))
		_, name, _ := splits[0], splits[1], splits[2]

		if _, ok := typeMap[name]; !ok {
			t.Errorf("TypeMap does not contain %s", name)
		}

		uniqueNames[name] = struct{}{}
	}

	if len(typeMap) > len(uniqueNames) {
		t.Errorf("typeMap contains too many types: %d > %d", len(typeMap), len(uniqueNames))

		// Print out the missing types
		for name := range uniqueNames {
			delete(typeMap, name)
		}

		for name := range typeMap {
			t.Errorf("typeMap contains unknown type %s", name)
		}
	}
}
