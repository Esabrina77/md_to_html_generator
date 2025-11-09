package main

import (
	"log"
	"os"
	"path/filepath"

	// IMPORTANT : Assurez-vous que "ssg-golang" correspond au nom de votre module
	"ssg-golang/internal/processor"
)

func main() {
	// Définition de nos dossiers sources et de sortie
	contentDir := "content"
	templateDir := "templates"
	outputDir := "public"

	log.Println("--- Démarrage du Générateur de Site Statique ---")

	// --- 1. Nettoyage du dossier de sortie ---
	if err := processor.SetupOutput(outputDir); err != nil {
		log.Fatalf("Échec de la préparation du dossier de sortie: %v", err)
	}

	// --- 2. Parcours des fichiers et déclenchement du traitement ---
	log.Printf("Recherche de fichiers dans: %s", contentDir)

	// La fonction filepath.Walk va parcourir récursivement 'contentDir'
	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Une erreur sur le parcours (dossier illisible, etc.)
			return err
		}

		// Nous traitons uniquement les fichiers (pas les dossiers) avec l'extension .md
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			// Appelle le cerveau du projet pour traiter le fichier trouvé
			if err := processor.ProcessFile(path, templateDir, outputDir); err != nil {
				log.Printf("ÉCHEC du traitement de %s: %v", path, err)
				// On log l'erreur mais on continue le parcours pour les autres fichiers
				return nil
			}
		}
		return nil // Continuer le parcours
	})

	if err != nil {
		log.Fatalf("Erreur fatale lors du parcours des fichiers: %v", err)
	}

	log.Println("--- Génération terminée avec succès! ---")
}
