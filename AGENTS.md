# Istruzioni per agenti: ClasseViva MCP

Questo file raccoglie le decisioni operative e i guasti già diagnosticati. Leggerlo prima di modificare il server, il plugin o la procedura d'installazione.

## Architettura di caricamento

- `.codex-plugin/plugin.json` installa la skill ClasseViva. Il server MCP viene registrato separatamente nel config locale dell'utente con il nome `classeviva`.
- `make install-local-plugin` compila `bin/classeviva`, aggiorna il marketplace locale, reinstalla la skill e registra il server MCP con il percorso assoluto del checkout. È la procedura standard dopo modifiche al server, agli strumenti o alla skill.
- `make install-local-mcp` ricompila e aggiorna soltanto la registrazione MCP. Usarlo quando serve riallineare il server senza reinstallare la skill.
- La registrazione MCP è locale all'utente e all'host Codex. Le nuove conversazioni locali, anche avviate in altre cartelle sullo stesso host, usano quel config. Ogni altro utente o host deve eseguire l'installazione nel proprio checkout. Un server stdio locale non è disponibile nelle conversazioni remote/cloud.
- Una conversazione già aperta non aggiorna il proprio elenco di strumenti. Dopo l'installazione verificare da una nuova conversazione locale.

## Errore di avvio già risolto

Una precedente configurazione avviava il server con `${PLUGIN_ROOT}/bin/classeviva` come campo `command`. Nel runtime usato per questa installazione il segnaposto è arrivato letteralmente a `execve`, che è fallito con `ENOENT`; il plugin appariva installato e abilitato, ma i suoi strumenti non venivano registrati. Anche la prova di spostare il segnaposto in `args` è rimasta letterale.

Per questo la configurazione attuale non dichiara un server bundled nel manifest: `make install-local-plugin` usa `codex mcp add` e salva un percorso assoluto del binario nel config dell'utente. Non reintrodurre `${PLUGIN_ROOT}` per avviare questo server senza prima aver verificato il comportamento del runtime con una sessione nuova e una chiamata MCP reale.

## Allegati della bacheca

- Non mescolare le API: la registrazione MCP attuale usa l'autenticazione mobile `/rest/v1`; il client web `/rest/w1` usa una sessione web distinta. Le fonti osservate descrivono parametri diversi per il download e un token mobile non ha funzionato su `/rest/w1`.
- L'endpoint mobile `GET /noticeboard/attach/{eventCode}/{pubId}/{attachNum}` richiede il valore `attachments[].attachNum` restituito dalla bacheca. Non usare un numero fisso come `101`: per avvisi reali con due allegati i valori `1` e `2` hanno restituito rispettivamente PDF e ZIP validi.
- Requisito di lettura osservato: su un avviso non letto, i GET con `attachNum` corretto hanno risposto HTTP 404 e `152:CvvRestApi/access to content not allowed`; dopo `POST /noticeboard/read/{eventCode}/{pubId}/101` (HTTP 200), gli stessi GET sono riusciti e hanno restituito i file. Per questo caso lo stato letto era quindi necessario; non è stata verificata una regola universale per tutti gli avvisi/account.
- `get_notice_attachment` non segna l'avviso come letto e al momento il server MCP non espone un tool per questa mutazione. Se il server restituisce `152:CvvRestApi/access to content not allowed`, spiega che la prova indica che potrebbe essere necessario segnare l'avviso come letto; il POST modifica lo stato e richiede consenso esplicito. Non ripetere il download sperando che l'errore scompaia e non aggiungere una marcatura automatica al tool di download. Se si vuole rendere questa azione disponibile agli agenti, aggiungere un tool separato con descrizione chiara dell'effetto.
- Prima di dichiarare riuscito un download, verificare che il file locale esista, sia non vuoto e abbia formato valido; per un tool MCP verificare anche tipo MIME e contenuto binario restituito.

## Verifica obbligatoria del caricamento

Un plugin elencato come `installed` e `enabled` non dimostra che il processo MCP sia partito o che i tool siano visibili. Controllare i passaggi separatamente:

1. Dopo l'installazione, verificare con `codex plugin list --json` che `classeviva` sia installato e abilitato.
2. Verificare con `codex mcp list --json` che il server `classeviva` sia abilitato e che `transport.command` sia un percorso assoluto al binario del checkout corrente.
3. Aprire una nuova sessione Codex locale e verificare che compaiano i tool MCP. Se possibile, ripetere da un altro repository per controllare che la registrazione non dipenda dalla cartella corrente.
4. Prima dei dati, chiamare `get_setup_status`. Se è configurato, verificare `list_lessons` e `list_absences` con un intervallo breve e riassumere soltanto stato e conteggio dei risultati.
5. I subagenti avviati con la delega nativa non ricevono l'inventario dei server MCP dell'utente: in una verifica, il subagente non aveva `get_setup_status` o `list_noticeboard` disponibili direttamente. Una sessione figlia avviata dal subagente con `codex exec --ephemeral --json` e accesso scrivibile allo stato locale (`$CODEX_HOME/state_*.sqlite`) ha invece caricato il server globale e chiamato entrambi i tool con successo. Un errore `readonly database` o `Read-only file system` è un blocco del sandbox prima dell'avvio MCP, non un errore del server. Non dichiarare verificata la delega finché una chiamata reale dalla sessione figlia non riesce.

`codex mcp list` mostra la configurazione, non prova da solo l'handshake MCP. In caso di dubbio raccogliere l'errore di avvio preciso. In passato `strace -f -e trace=execve` ha permesso di confermare che il processo riceveva il segnaposto letterale.

Una sessione CLI avviata in sandbox di sola lettura può fallire prima dell'MCP perché Codex non riesce a scrivere il proprio database di sessione sotto `CODEX_HOME`. Non confondere questo errore con un guasto del server: ripetere la verifica in un runtime locale autorizzato a scrivere il proprio stato.

## Modifiche al server e rilascio

- Quando si aggiunge o modifica un tool, aggiornare anche `skills/classeviva-setup/SKILL.md` e la tabella degli strumenti in `README.md`, mantenendo coerenti nomi, parametri e intervalli data.
- Tenere allineata la versione in `.codex-plugin/plugin.json` con quella dichiarata in `cmd/classeviva/main.go`; ricompilare il binario tracciato `bin/classeviva`.
- Dopo modifiche al codice, ai tool o alla skill, eseguire `make install-local-plugin` per aggiornare la copia locale e la registrazione usate dalle conversazioni future.
- Prima di pubblicare, verificare il diff e l'esito di `go test ./...`; poi controllare installazione, registrazione MCP e almeno una chiamata reale da una sessione nuova.

## Dati scolastici e credenziali

- Non leggere o stampare il file delle credenziali e non riportare UID o password. Usare `get_setup_status`, che restituisce solo lo stato e il percorso.
- Non salvare credenziali o risposte personali di ClasseViva nel repository, nei log condivisi, nei test o in questo file.
- Durante le verifiche limitare date e quantità di dati; riportare esiti e conteggi, non materie, docenti, assenze o altri dettagli personali.
- Le chiamate di consultazione sono in sola lettura. Per scaricare materiale didattico, farlo solo su richiesta esplicita e nel percorso indicato dall'utente.
