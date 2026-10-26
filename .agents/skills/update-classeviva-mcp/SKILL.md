---
name: update-classeviva-mcp
description: Use when the user asks to update, upgrade, refresh, pull, or reinstall this ClasseViva repository, MCP server, or Codex plugin. Finds the active checkout, advances it safely, reinstalls the local plugin and server, and preserves machine-local credentials and files.
---

# Aggiornare ClasseViva MCP e il plugin

Questa procedura aggiorna il clone Git da cui è installato il plugin locale, ricompila il server per il computer corrente e reinstalla plugin e registrazione MCP. Il marketplace `classeviva-repo` punta a un checkout locale: aggiornare il catalogo del marketplace non scarica i nuovi commit del repository.

## 1. Individua il checkout installato

Se l'utente ha indicato la cartella del progetto, verifica che sia il checkout effettivamente usato. Altrimenti:

1. Esegui `codex plugin marketplace list` e individua il percorso locale del marketplace `classeviva-repo`.
2. Controlla `codex plugin list --json` e `codex mcp list --json`. La registrazione MCP `classeviva` deve puntare al binario `bin/classeviva` dentro lo stesso checkout.
3. Verifica la radice con `git -C <checkout> rev-parse --show-toplevel` e che il remoto atteso sia `github.com/Maro314o/classeviva-mcp`.

Non aggiornare un clone diverso da quello collegato all'installazione. Se ci sono più checkout e non si può stabilire quale sia quello attivo, chiedi all'utente quale aggiornare.

## 2. Proteggi modifiche e configurazione locale

Prima di cambiare file, controlla:

```sh
git -C <checkout> status --short --branch
git -C <checkout> branch --show-current
```

Procedi automaticamente solo se il checkout è pulito e si trova su `main`. Se ci sono modifiche tracciate, commit locali, un branch diverso o un aggiornamento non fast-forward, fermati prima di modificare il repository. Non usare `git reset`, `git clean`, `git stash`, rebase o merge forzati per superare il blocco: descrivi cosa hai trovato e concorda come preservarlo.

Le credenziali non sono nel checkout: il server le conserva nel file `classeviva/credentials.txt` nella cartella di configurazione dell'utente del sistema operativo, salvo override `CLASSEVIVA_CREDENTIALS_FILE`. Non aprire, stampare, copiare o cancellare quel file. I download nella cartella `downloads/` del repository sono ignorati da Git; non usare comandi di pulizia che li rimuovano. Un altro computer ha una propria configurazione e, se necessario, va configurato dal browser locale; non copiare le credenziali tra computer.

`make install-local-plugin` reinstalla il plugin `classeviva` e rimuove/ricrea la registrazione MCP omonima. Prima di eseguirlo, controlla con `codex mcp list --json` che quella voce punti al checkout e non abbia argomenti, variabili d'ambiente o altre opzioni personalizzate che andrebbero perse. Se le trovi, fermati e preservale concordando una modifica adatta alla configurazione esistente.

## 3. Porta il checkout all'ultima versione

Dal checkout pulito su `main`, aggiorna con fast-forward soltanto:

```sh
git -C <checkout> fetch origin
git -C <checkout> pull --ff-only origin main
```

Se il fetch fallisce, riferisci l'errore e non simulare l'aggiornamento. Se il pull non può avanzare in fast-forward, non forzare: conserva il branch e i commit locali e chiedi come procedere.

## 4. Ricompila e reinstalla sul computer corrente

Verifica la disponibilità di Go 1.25 o successivo e di `make`, poi esegui dalla radice del checkout:

```sh
make install-local-plugin
```

Il target compila il binario per il sistema operativo e l'architettura correnti, aggiorna il marketplace locale, reinstalla il plugin e registra il percorso assoluto del binario nel config utente di Codex. È il passaggio che applica localmente sia il codice appena scaricato sia le istruzioni aggiornate della skill. Non richiede di toccare credenziali o download.

## 5. Verifica e riapri la sessione

Controlla dopo l'installazione:

```sh
codex plugin list --json
codex mcp list --json
git -C <checkout> status --short --branch
```

Verifica che `classeviva` sia installato e abilitato e che il comando MCP sia il percorso assoluto del binario appena compilato in quel checkout. La compilazione è eseguita da `make install-local-plugin`; se il target fallisce, non dichiarare l'update completato.

Le conversazioni già aperte non ricaricano l'elenco di strumenti. Chiedi all'utente di aprire una nuova conversazione locale, controlla che i tool MCP siano visibili e chiama solo `get_setup_status` per verificare che la configurazione sia ancora disponibile. Se restituisce `configured: false`, avvia la normale configurazione nel browser locale; non leggere né trasferire credenziali. Non consultare dati scolastici come prova dell'aggiornamento.

## Cosa riferire

Riassumi il commit raggiunto, l'esito della compilazione/reinstallazione e la verifica della configurazione MCP. Distingui sempre tra plugin reinstallato e strumenti effettivamente visti in una nuova conversazione: `codex plugin list` e `codex mcp list` verificano la configurazione, ma la visibilità dei tool si conferma solo dopo aver avviato una nuova sessione.
