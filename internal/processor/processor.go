package processor

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/gomarkdown/markdown"
)

// Metadata représente les données extraites du Front Matter YAML.
type Metadata struct {
	// Le tag `yaml:"title"` est crucial pour que le package yaml.v3
	// sache où placer la valeur lue.
	Title string `yaml:"title"`
}

// PageData est la structure qui contient les données
// qui seront injectées dans nos modèles HTML.
type PageData struct {
	Metadata Metadata // Contient le Titre
	Content  template.HTML
}

// SetupOutput prépare le dossier de sortie.
// Il le supprime s'il existe pour garantir une nouvelle génération.
func SetupOutput(dir string) error {
	log.Printf("Nettoyage du dossier de sortie: %s", dir)
	// 1. Supprimer le dossier (y compris son contenu)
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("erreur lors de la suppression de %s: %w", dir, err)
	}
	// 2. Recréer le dossier
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("erreur lors de la création de %s: %w", dir, err)
	}
	return nil
}

// ProcessFile est le cœur de notre générateur.
// Il lit un fichier, le convertit, extrait les métadonnées et fusionne avec les modèles.
func ProcessFile(filePath, templateDir, outputDir string) error {
	// 1. Lire le contenu brut du fichier
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("erreur de lecture de %s: %w", filePath, err)
	}

	content := string(contentBytes)

	// --- A. Extraction du Front Matter ---

	// On cherche les délimiteurs du Front Matter (---)
	yamlStart := strings.Index(content, "---")
	yamlEnd := strings.Index(content[yamlStart+3:], "---")

	var yamlMetadata string
	var markdownContent []byte

	if yamlStart == 0 && yamlEnd > 0 {
		// Le Front Matter est présent et correct (commence et se termine par ---)
		yamlMetadata = content[yamlStart+3 : yamlEnd+3] // +3 pour inclure les délimiteurs
		markdownContent = contentBytes[yamlEnd+6:]      // +6 pour sauter les deux "---" et les retours à la ligne
	} else {
		// Pas de Front Matter trouvé, on considère tout comme du Markdown
		markdownContent = contentBytes
		log.Printf("Avertissement: Pas de Front Matter trouvé dans %s. Titre par défaut.", filePath)
	}

	// 2. Analyser le YAML
	var metadata Metadata
	if yamlMetadata != "" {
		// Utilisation du package yaml.v3 pour décoder les données dans la structure Metadata
		if err := yaml.Unmarshal([]byte(yamlMetadata), &metadata); err != nil {
			return fmt.Errorf("erreur de parsing YAML dans %s: %w", filePath, err)
		}
	}

	// --- B. Conversion du Markdown en HTML ---

	// Conversion du Markdown restant en HTML brut
	htmlContent := markdown.ToHTML(markdownContent, nil, nil)

	// --- C. Préparation des Modèles et des Données ---
	tmpl, err := template.ParseFiles(
		filepath.Join(templateDir, "base.html"),
		filepath.Join(templateDir, "page.html"),
	)
	if err != nil {
		return fmt.Errorf("erreur de parsing des modèles: %w", err)
	}

	// 3. Préparer les données pour l'injection (ENFIN DYNAMIQUE !)
	data := PageData{
		Metadata: metadata,
		Content:  template.HTML(htmlContent),
	}

	// --- D. Détermination du Chemin de Sortie et Écriture ---
	relPath, _ := filepath.Rel("content", filePath)
	outPath := filepath.Join(outputDir, relPath)
	outPath = strings.Replace(outPath, ".md", ".html", 1)

	if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("erreur de création de %s: %w", outPath, err)
	}
	defer f.Close()

	// 4. Exécuter le modèle.
	if err := tmpl.ExecuteTemplate(f, "base", data); err != nil {
		return fmt.Errorf("erreur d'exécution du modèle: %w", err)
	}

	log.Printf("Généré avec succès: %s (Titre: %s)", outPath, metadata.Title)
	return nil
}
