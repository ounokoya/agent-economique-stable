# Workflows SDK BingX - Documentation Détaillée

## 📋 Vue d'Ensemble

Documentation complète des workflows pour le SDK BingX en Go, couvrant toutes les actions trading fondamentales sur Spot et Futures Perpétuels avec gestion multi-comptes.

---

## 🔄 WORKFLOWS SPOT TRADING

### 🟢 Workflow Achat Spot (BUY)

#### Étapes Détaillées :

1. **Initialisation Environnement**
   - Validation environnement (Demo VST / Live Prod)
   - Chargement configuration API appropriée
   - Vérification connectivité réseau

2. **Authentification Client**
   - Génération signature HMAC SHA256
   - Validation API keys et permissions
   - Test connexion avec endpoint de santé

3. **Validation Solde Disponible**
   - Récupération solde USDT sur compte/sous-compte
   - Vérification montant suffisant pour achat
   - Calcul frais estimés de transaction

4. **Récupération Prix Marché**
   - Appel endpoint prix temps réel
   - Validation cohérence prix avec carnet ordres
   - Application rate limiting (1 req/sec max)

5. **Calcul et Validation Quantité**
   - Détermination quantité crypto selon budget
   - Validation minimums/maximums exchange
   - Ajustement selon précision symbole

6. **Préparation Ordre**
   - Construction paramètres ordre (Market/Limit)
   - Validation tous champs requis
   - Calcul frais finaux estimés

7. **Placement Ordre**
   - Envoi requête POST vers endpoint trading
   - Gestion erreurs et retry si nécessaire
   - Récupération ID ordre pour suivi

8. **Surveillance Statut**
   - Polling statut ordre jusqu'à exécution
   - Gestion timeout et annulation si besoin
   - Logging détaillé de chaque étape

9. **Mise à Jour Post-Exécution**
   - Récupération détails exécution finale
   - Mise à jour soldes locaux
   - Calcul prix d'exécution réel

10. **Logging et Audit**
    - Enregistrement transaction complète
    - Mise à jour historique trades
    - Notification système si configurée

#### Points de Contrôle :
- Validation solde avant placement
- Vérification prix cohérent
- Confirmation exécution complète
- Audit trail complet

#### Gestion d'Erreurs :
- Solde insuffisant → Exception avec détails
- Prix invalide → Retry avec nouveau prix
- Timeout ordre → Annulation automatique
- Erreur réseau → Retry avec backoff

---

### 🔴 Workflow Vente Spot (SELL)

#### Étapes Détaillées :

1. **Préparation Environnement**
   - Configuration environnement (Demo/Live)
   - Authentification avec API keys
   - Validation permissions trading

2. **Validation Possession Crypto**
   - Vérification solde crypto disponible
   - Contrôle quantité minimum vendable
   - Validation pas de crypto en ordre ouvert

3. **Analyse Prix Vente**
   - Récupération prix bid actuel
   - Évaluation spread bid/ask
   - Estimation montant USDT final

4. **Calcul Montant Final**
   - Soustraction frais de trading
   - Application taux commission
   - Validation montant minimum réception

5. **Configuration Ordre Vente**
   - Choix type ordre (Market/Limit)
   - Définition prix si Limit
   - Validation paramètres complets

6. **Exécution Vente**
   - Placement ordre sur exchange
   - Monitoring progression exécution
   - Gestion fills partiels si applicable

7. **Confirmation et Finalisation**
   - Vérification exécution complète
   - Récupération USDT reçus
   - Validation montants attendus

8. **Mise à Jour Système**
   - Update soldes crypto et USDT
   - Enregistrement dans historique
   - Calcul PnL si achat initial connu

9. **Notification et Logging**
   - Log transaction avec tous détails
   - Notification résultat si configurée
   - Mise à jour métriques performance

#### Points de Validation :
- Possession suffisante crypto
- Prix vente acceptable
- Frais correctement calculés
- USDT reçus conformes

---

## ⚡ WORKFLOWS FUTURES TRADING

### 📈 Workflow Ouverture Position Long

#### Étapes Détaillées :

1. **Configuration Trading**
   - Sélection environnement (Demo/Live)
   - Configuration levier souhaité (1x-125x)
   - Choix mode marge (Cross/Isolated)

2. **Validation Paramètres**
   - Vérification levier autorisé pour symbole
   - Validation mode marge supporté
   - Contrôle taille position dans limites

3. **Calcul Marge Requise**
   - Calcul marge initiale selon levier
   - Ajout buffer pour fluctuations
   - Vérification solde suffisant

4. **Analyse Risque Position**
   - Calcul prix liquidation théorique
   - Évaluation exposition maximale
   - Validation ratio risque/capital

5. **Récupération Prix Entrée**
   - Prix mark actuel du contrat
   - Analyse spread et slippage potentiel
   - Optimisation timing d'entrée

6. **Placement Ordre Ouverture**
   - Construction ordre Long (BUY)
   - Spécification positionSide si Hedge
   - Envoi vers endpoint trading

7. **Confirmation Position Ouverte**
   - Vérification position dans portefeuille
   - Validation paramètres (taille, levier, marge)
   - Récupération prix d'entrée réel

8. **Initialisation Monitoring**
   - Calcul PnL initial (généralement 0)
   - Configuration trailing stop si activé
   - Setup surveillance continue

9. **Activation Gestion Risque**
   - Définition stop loss si configuré
   - Activation take profit si défini
   - Initialisation ajustements dynamiques

10. **Logging et Suivi**
    - Enregistrement ouverture position
    - Début monitoring PnL temps réel
    - Notification ouverture si configurée

#### Indicateurs Intégrés :
- **MACD** : Signal haussier confirmé avant ouverture
- **CCI** : Zone survente validée
- **DMI** : Tendance ou contre-tendance selon configuration

---

### 📉 Workflow Ouverture Position Short

#### Étapes Détaillées :

1. **Préparation Environnement**
   - Configuration environnement trading
   - Authentification permissions futures
   - Validation short autorisé sur symbole

2. **Configuration Paramètres Short**
   - Définition levier approprié
   - Sélection mode marge optimal
   - Validation limites position short

3. **Analyse Conditions Marché**
   - Vérification tendance baissière
   - Analyse volume et volatilité
   - Évaluation timing d'entrée

4. **Calcul Exposition Risque**
   - Marge requise pour position short
   - Prix liquidation côté hausse
   - Impact funding rate sur position

5. **Validation Stratégique**
   - Confirmation signaux baissiers
   - Validation absence résistances majeures
   - Contrôle corrélations marché

6. **Exécution Ordre Short**
   - Placement ordre SELL pour ouverture
   - Spécification SHORT en positionSide
   - Monitoring exécution temps réel

7. **Confirmation Position Short**
   - Vérification position négative en portefeuille
   - Contrôle paramètres configuration
   - Validation marge allouée correctement

8. **Setup Surveillance**
   - Initialisation monitoring PnL
   - Configuration stops adaptatifs
   - Activation alertes mouvement adverse

9. **Gestion Continue**
   - Surveillance prix en temps réel
   - Ajustements trailing stop
   - Monitoring funding rate impact

#### Signaux Déclencheurs :
- **MACD** : Croisement baissier confirmé
- **CCI** : Zone surachat validée  
- **DMI** : Momentum baissier selon tendance

---

### ✅ Workflow Fermeture Positions

#### Processus Unifié Long/Short :

1. **Identification Position**
   - Récupération position ciblée
   - Validation état ouvert actuel
   - Contrôle permissions fermeture

2. **Évaluation PnL Courant**
   - Calcul profit/perte temps réel
   - Impact frais de fermeture
   - Évaluation opportunity cost

3. **Décision Fermeture**
   - Analyse signaux sortie
   - Validation conditions trailing stop
   - Confirmation critères stratégiques

4. **Choix Type Fermeture**
   - Fermeture totale vs partielle
   - Ordre Market vs Limit
   - Urgence vs optimisation prix

5. **Préparation Ordre Fermeture**
   - Calcul quantité exacte à fermer
   - Inversion side original (SELL→BUY, BUY→SELL)
   - Configuration reduce-only

6. **Exécution Fermeture**
   - Placement ordre fermeture
   - Monitoring exécution progressive
   - Gestion fills partiels

7. **Confirmation Fermeture**
   - Vérification position fermée/réduite
   - Récupération détails exécution
   - Validation cohérence quantités

8. **Calcul PnL Final**
   - Calcul profit/perte réalisé
   - Déduction frais trading
   - Impact funding rate total

9. **Libération Marge**
   - Retour marge au solde disponible
   - Recalcul pouvoir d'achat
   - Validation soldes cohérents

10. **Finalisation**
    - Enregistrement transaction close
    - Mise à jour historique positions
    - Notification et reporting

#### Conditions de Sortie :
- **Trailing stop** déclenché
- **Signal inverse** MACD/CCI/DMI
- **Take profit** atteint
- **Stop loss** touché
- **Fermeture manuelle** utilisateur

---

## 🏦 WORKFLOWS MULTI-COMPTES

### 💰 Workflow Gestion Sous-Comptes

#### Cycle Complet :

1. **Création Sous-Compte**
   - Génération nom unique sous-compte
   - Appel API création BingX
   - Récupération UID sous-compte

2. **Configuration API Keys**
   - Génération API key dédiée
   - Configuration permissions granulaires
   - Test authentification nouvelle key

3. **Allocation Budget Initial**
   - Calcul budget selon stratégie
   - Transfert depuis compte principal
   - Validation transfert effectué

4. **Initialisation Bot**
   - Association bot à sous-compte
   - Configuration paramètres trading
   - Démarrage monitoring activité

5. **Surveillance Continue**
   - Monitoring performance temps réel
   - Suivi utilisation budget
   - Alertes dépassements limites

6. **Gestion Profits**
   - Récupération profits périodique
   - Transfert vers compte principal
   - Conservation capital de travail

7. **Rebalancement**
   - Analyse performance relative
   - Réallocation budgets si nécessaire
   - Optimisation distribution capital

8. **Audit et Conformité**
   - Vérification permissions API
   - Contrôle transactions suspectes
   - Mise à jour sécurité

#### Isolation Risques :
- **Séparation complète** fonds par bot
- **Permissions limitées** par API key
- **Monitoring centralisé** sans accès croisé
- **Audit trail** complet par sous-compte

---

## 🔧 WORKFLOWS TECHNIQUES

### ⚙️ Workflow Rate Limiting

#### Gestion Intelligente :

1. **Initialisation Limiters**
   - Configuration limits par type endpoint
   - Setup burst allowance temporaire
   - Initialisation compteurs requêtes

2. **Classification Requêtes**
   - Market data (10 req/sec max)
   - Trading orders (priorité haute)
   - Account info (priorité moyenne)

3. **Distribution Intelligente**
   - Répartition sur multiple sous-comptes
   - Load balancing entre serveurs
   - Optimisation timing requêtes

4. **Monitoring et Ajustement**
   - Surveillance erreurs 429
   - Ajustement dynamique limits
   - Reporting utilisation quotidienne

---

### 🔄 Workflow Cache et Performance

#### Optimisation Données :

1. **Cache Prix Temps Réel**
   - TTL court (5-10 secondes)
   - Invalidation sur ordre exécuté
   - Partage entre bots même symbole

2. **Cache Informations Statiques**
   - Symboles et configurations
   - TTL long (1 heure)
   - Refresh périodique background

3. **Batch Operations**
   - Regroupement requêtes compatibles
   - Optimisation bande passante
   - Réduction latence globale

---

## 📊 INTÉGRATION STRATÉGIES MACD/CCI/DMI

### 🎯 Workflow Signaux Trading

#### Pipeline Décisionnel :

1. **Collecte Données Indicateurs**
   - Récupération klines historiques
   - Calcul MACD, CCI, DMI temps réel
   - Validation cohérence données

2. **Analyse Signaux**
   - Détection croisements MACD
   - Identification zones extrêmes CCI
   - Évaluation force tendance DMI

3. **Génération Recommandations**
   - Combinaison signaux selon règles
   - Calcul niveau confidence
   - Filtrage signaux faibles

4. **Exécution Conditionnelle**
   - Validation confidence > seuil (0.7)
   - Ouverture position si conditions remplies
   - Ajustements trailing stop dynamiques

5. **Monitoring Continu**
   - Surveillance signaux inverses
   - Ajustements position en cours
   - Déclenchement sorties anticipées

#### Points de Contrôle Stratégiques :
- **Signal MACD** : Croisement confirmé
- **Zone CCI** : Extrême validé (>100 ou <-100)
- **Tendance DMI** : Direction cohérente
- **Confidence globale** : Score > 70%

---

## 🎯 Conclusion Workflows

**Architecture complète** couvrant tous les aspects trading automatisé sur BingX.

**Intégration native** avec stratégies MACD/CCI/DMI existantes.

**Gestion multi-comptes** robuste avec isolation risques.

**Performance optimisée** respectant rate limits et contraintes.

**Prêt pour implémentation** avec workflows détaillés et points de contrôle précis.
