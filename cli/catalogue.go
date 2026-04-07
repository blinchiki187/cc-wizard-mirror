package cli
import (
	"log"
	"fmt"
	"net/http"
	"slices"
	"maps"
	"encoding/json"
	"coopcloud.tech/abra/pkg/recipe"

)

func (h *abraHandler) handleListCatalogue(w http.ResponseWriter, r *http.Request) {
	offline := false
	catl, err := recipe.ReadRecipeCatalogue(offline)
	if err != nil {
		log.Fatal(err)
	}
    vals := slices.Collect(maps.Values(catl))

	jsonBytes, err := json.Marshal(vals)
	if err != nil {
		log.Printf("Error: %s\n", err)
		http.Error(w, fmt.Sprintf("Error: %s\n", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

