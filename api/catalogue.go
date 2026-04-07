package api
import (
	"context"
	"fmt"
	"log"
	"encoding/json"
	"coopcloud.tech/abra/pkg/client"
	"coopcloud.tech/abra/pkg/upstream/stack"
	"coopcloud.tech/abra/pkg/upstream/convert"
	"coopcloud.tech/abra/pkg/recipe"

	"coop-cloud-backend/internal"
)

func (h *abraHandler) handleListCatalogue(w http.ResponseWriter, r *http.Request) {
	catalogue, err := recipe.ReadRecipeCatalogue(internal.Offline)
	if err != nil {
		log.Fatal(err)
	}

	recipes := catalogue.Flatten()

	jsonBytes, err := json.Marshal(recipes)
	if err != nil {
		log.Printf("JSON conversion failed: %s\n", err)
		http.Error(w, fmt.Sprintf("JSON conversion failed: %s\n", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

