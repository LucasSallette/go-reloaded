package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("il faut faire go run main.go <fichier_entree> <fichier_sortie>")
		os.Exit(1)
	}

	fichierEntree := os.Args[1]
	fichierSortie := os.Args[2]

	contenu, err := os.ReadFile(fichierEntree)
	if err != nil {
		os.Exit(1)
	}

	resultat := traiter(string(contenu))

	err = os.WriteFile(fichierSortie, []byte(resultat), 0644)
	if err != nil {
		os.Exit(1)
	}
}

// traiter applique toutes les transformations sur le texte
func traiter(texte string) string {
	// Découper le texte en mots
	mots := strings.Fields(texte)

	// Fusionner les commandes en deux parties ex: "(cap," "3)" → "(cap, 3)"
	mots = fusionnerCommandesNombrees(mots)

	// Appliquer les commandes (hex), (bin), (up), (low), (cap)
	mots = appliquerTransformations(mots)

	// Remettre les mots ensemble
	resultat := strings.Join(mots, " ")

	// Corriger la ponctuation (espaces avant . , ! ? : ; et apostrophes)
	resultat = corrigerPonctuation(resultat)

	// Corriger "a" → "an" devant une voyelle
	resultat = corrigerArticleA(resultat)

	return resultat
}

// fusionnerCommandesNombrees colle les tokens "(cap," et "3)" en un seul "(cap, 3)"
// car strings.Fields sépare les espaces à l'intérieur des parenthèses
func fusionnerCommandesNombrees(mots []string) []string {
	resultat := []string{}

	for i := 0; i < len(mots); i++ {
		mot := mots[i]

		// Si le mot commence par "(" et finit par "," sans ")" → c'est une commande coupée
		if strings.HasPrefix(mot, "(") && strings.HasSuffix(mot, ",") && !strings.Contains(mot, ")") {
			// Le mot suivant devrait être le nombre avec la parenthèse fermante
			if i+1 < len(mots) && strings.HasSuffix(mots[i+1], ")") {
				motFusionne := mot + " " + mots[i+1]
				resultat = append(resultat, motFusionne)
				i++ // on saute le mot suivant car il est déjà fusionné
				continue
			}
		}

		resultat = append(resultat, mot)
	}

	return resultat
}

// appliquerTransformations parcourt les mots et applique les commandes
func appliquerTransformations(mots []string) []string {
	resultat := []string{}

	for i := 0; i < len(mots); i++ {
		mot := mots[i]

		// Commande simple : (hex), (bin), (up), (low), (cap)
		if estCommandeSimple(mot) {
			commande := mot[1 : len(mot)-1] // enlever les parenthèses
			commande = strings.ToLower(commande)

			// Appliquer la commande sur le dernier mot ajouté
			if len(resultat) > 0 {
				resultat[len(resultat)-1] = appliquerCommande(resultat[len(resultat)-1], commande)
			}
			continue
		}

		// Commande avec nombre : (up, 3), (low, 2), (cap, 5)
		if estCommandeNombree(mot) {
			commande, nombre := extraireCommandeNombree(mot)

			// Appliquer la commande sur les N derniers mots
			debut := len(resultat) - nombre
			if debut < 0 {
				debut = 0
			}
			for j := debut; j < len(resultat); j++ {
				resultat[j] = appliquerCommande(resultat[j], commande)
			}
			continue
		}

		// Mot normal, on l'ajoute tel quel
		resultat = append(resultat, mot)
	}

	return resultat
}

// estCommandeSimple vérifie si le mot est une commande comme (hex), (bin), (up), (low), (cap)
func estCommandeSimple(mot string) bool {
	motMinuscule := strings.ToLower(mot)
	return motMinuscule == "(hex)" ||
		motMinuscule == "(bin)" ||
		motMinuscule == "(up)" ||
		motMinuscule == "(low)" ||
		motMinuscule == "(cap)"
}

// estCommandeNombree vérifie si le mot est une commande avec nombre ex: "(cap, 3)"
func estCommandeNombree(mot string) bool {
	// Le mot doit commencer par "(" et finir par ")"
	if !strings.HasPrefix(mot, "(") || !strings.HasSuffix(mot, ")") {
		return false
	}

	// On enlève les parenthèses et on sépare par la virgule
	interieur := mot[1 : len(mot)-1]
	parties := strings.SplitN(interieur, ",", 2)
	if len(parties) != 2 {
		return false
	}

	commande := strings.ToLower(strings.TrimSpace(parties[0]))
	nombre := strings.TrimSpace(parties[1])

	// Vérifier que c'est une commande valide et que le nombre est un entier
	_, erreur := strconv.Atoi(nombre)
	return erreur == nil && (commande == "up" || commande == "low" || commande == "cap")
}

// extraireCommandeNombree retourne la commande et le nombre depuis "(cap, 3)"
func extraireCommandeNombree(mot string) (string, int) {
	interieur := mot[1 : len(mot)-1]
	parties := strings.SplitN(interieur, ",", 2)
	commande := strings.ToLower(strings.TrimSpace(parties[0]))
	nombre, _ := strconv.Atoi(strings.TrimSpace(parties[1]))
	return commande, nombre
}

// appliquerCommande applique une transformation sur un mot
func appliquerCommande(mot string, commande string) string {
	if commande == "up" {
		// Mettre en majuscules
		return strings.ToUpper(mot)
	}

	if commande == "low" {
		// Mettre en minuscules
		return strings.ToLower(mot)
	}

	if commande == "cap" {
		// Mettre la première lettre en majuscule, le reste en minuscule
		return mettreEnMajuscule(mot)
	}

	if commande == "hex" {
		// Convertir un nombre hexadécimal en décimal
		nombre, erreur := strconv.ParseInt(mot, 16, 64)
		if erreur != nil {
			return mot // si erreur, on retourne le mot sans changement
		}
		return strconv.FormatInt(nombre, 10)
	}

	if commande == "bin" {
		// Convertir un nombre binaire en décimal
		nombre, erreur := strconv.ParseInt(mot, 2, 64)
		if erreur != nil {
			return mot // si erreur, on retourne le mot sans changement
		}
		return strconv.FormatInt(nombre, 10)
	}

	return mot
}

// mettreEnMajuscule met la première lettre en majuscule et le reste en minuscule
func mettreEnMajuscule(mot string) string {
	if len(mot) == 0 {
		return mot
	}

	// Convertir en slice de runes pour bien gérer les caractères spéciaux
	lettres := []rune(mot)
	lettres[0] = unicode.ToUpper(lettres[0])
	for i := 1; i < len(lettres); i++ {
		lettres[i] = unicode.ToLower(lettres[i])
	}

	return string(lettres)
}

// corrigerPonctuation corrige les espaces autour de la ponctuation
// Règle : pas d'espace AVANT . , ! ? : ;
// Règle : pas d'espace autour des apostrophes '
func corrigerPonctuation(texte string) string {
	// Supprimer les espaces en double
	texte = strings.Join(strings.Fields(texte), " ")

	// Supprimer l'espace avant chaque signe de ponctuation
	signes := []string{".", ",", "!", "?", ":", ";"}
	for _, signe := range signes {
		texte = strings.ReplaceAll(texte, " "+signe, signe)
	}

	// Coller les apostrophes (supprimer les espaces autour)
	texte = corrigerApostrophes(texte)

	return texte
}

// corrigerApostrophes supprime les espaces autour des apostrophes
// Exemple : "don ' t" → "don't"
func corrigerApostrophes(texte string) string {
	lettres := []rune(texte)
	longueur := len(lettres)
	resultat := ""

	i := 0
	for i < longueur {
		if lettres[i] == '\'' {
			// Supprimer l'espace qui précède l'apostrophe
			resultat = strings.TrimRight(resultat, " ")
			resultat += "'"

			// Avancer en sautant les espaces qui suivent l'apostrophe
			i++
			for i < longueur && lettres[i] == ' ' {
				i++
			}
			continue
		}

		resultat += string(lettres[i])
		i++
	}

	return resultat
}

// corrigerArticleA remplace "a" par "an" quand le mot suivant commence par une voyelle
// Exemple : "a apple" → "an apple"
func corrigerArticleA(texte string) string {
	voyelles := "aeiouAEIOU"
	mots := strings.Split(texte, " ")

	for i := 0; i < len(mots)-1; i++ {
		// Vérifier si le mot est l'article "a" ou "A"
		if mots[i] == "a" || mots[i] == "A" {
			motSuivant := mots[i+1]

			// Trouver la première lettre du mot suivant
			premiereLetttre := premiereLettredu(motSuivant)

			// Si cette lettre est une voyelle, changer "a" en "an"
			if premiereLetttre != 0 && strings.ContainsRune(voyelles, premiereLetttre) {
				if mots[i] == "A" {
					mots[i] = "An"
				} else {
					mots[i] = "an"
				}
			}
		}
	}

	return strings.Join(mots, " ")
}

// premiereLettredu retourne la première lettre d'un mot (ignore la ponctuation)
func premiereLettredu(mot string) rune {
	for _, lettre := range mot {
		if unicode.IsLetter(lettre) {
			return lettre
		}
	}
	return 0
}
