# Barb Quickstart

This guide covers connecting Barb to the Mirage quickstart environment so you can run a full phishing campaign end-to-end using the bundled target site.

## Prerequisites

Complete the [Mirage Quickstart](https://github.com/travisbale/mirage/blob/master/docs/quickstart.md) first. You should have:

- `miraged` and the target site running via Docker Compose
- `/etc/hosts` entries for `*.phish.local` and `*.target.local`
- The self-signed CA trusted in your browser

## 1. Build and start Barb

```bash
make build
./build/barb serve --addr :4443 --debug
```

Port 4443 avoids conflicting with miraged which runs on 443. Open `https://localhost:4443` in your browser. On first login, Barb will prompt you to set a password.

## 2. Connect to miraged

Navigate to **Miraged** and add a new miraged connection:

1. Find the enrollment token in the miraged logs:

   ```bash
   docker compose -f /path/to/mirage/examples/quickstart/docker-compose.yml logs miraged | grep "enroll with"
   ```

2. In Barb's connection form, enter:
   - **Address:** `127.0.0.1:443`
   - **Secret hostname:** `mgmt.phish.local`
   - **Token:** the token from the logs

Barb generates a keypair and enrolls with miraged automatically.

### Optional: add notification channels

After enrollment, Barb opens the connection's detail page. Use the **Notification Channels** section to forward miraged events (`session.created`, `creds.captured`, etc.) to a webhook or Slack incoming webhook URL. Leave the event filter empty to receive all events, or pick specific types. The **Test** button sends a sample event to verify delivery.

## 3. Set up an SMTP profile

For local testing, use [Mailpit](https://mailpit.axigen.com/) as a fake SMTP server:

```bash
docker run -d --name mailpit -p 1025:1025 -p 8025:8025 \
  -e MP_SMTP_AUTH_ACCEPT_ANY=1 \
  -e MP_SMTP_AUTH_ALLOW_INSECURE=1 \
  axllent/mailpit:latest
```

In Barb, navigate to **SMTP Profiles** and create a profile:

- **Host:** `localhost`
- **Port:** `1025`
- **From:** `security@phish.local`

Mailpit's web UI at `http://localhost:8025` shows all delivered emails.

## 4. Upload a phishlet

Navigate to **Phishlets** and create a new phishlet. Paste or upload the contents of one of the example phishlet files from the mirage repo (e.g., `examples/phishlets/form-login.yaml`).

## 5. Create a target list

Navigate to **Target Lists** and create a list with at least one target. For local testing, any email address works — Mailpit accepts everything.

## 6. Create an email template

Navigate to **Templates** and create an email template. The template body supports Go template variables:

- `{{.FirstName}}` — target's first name
- `{{.LastName}}` — target's last name
- `{{.Email}}` — target's email address
- `{{.URL}}` — the unique lure URL (generated per-target with encrypted tracking parameters)

Include `{{.URL}}` as a link in the body — this is what the target clicks to reach the phishing proxy.

## 7. Launch a campaign

Navigate to **Campaigns** and create a new campaign using the wizard:

1. Select the miraged connection
2. Select the phishlet and set the hostname to `login.phish.local`
3. Select the target list
4. Select the email template and SMTP profile
5. Set the redirect URL to `https://login.target.local:8443/demo-complete`
6. Start the campaign

Barb pushes the phishlet to miraged, creates a lure, and begins sending emails. The campaign detail page updates in real time as emails are sent, links are clicked, and sessions are captured — no need to refresh.

Check Mailpit at `http://localhost:8025` to see the delivered emails. Each email contains a unique lure URL with an encrypted tracking parameter. Click the link to walk through the login flow on the proxied target site.

As targets interact with the phishing site, the campaign results progress through the lifecycle: **sent** → **clicked** (target visited the lure) → **captured** (credentials submitted) → **completed** (full session with auth tokens captured).

## 8. View captured sessions

Click a completed result in the campaign detail view to see the captured credentials, auth tokens, and session metadata. Use the **Export Cookies** button to download session cookies for browser import, or **Export CSV** to download campaign results for reporting.

## Next steps

- Try different phishlet configurations for other target applications
- Test with multiple targets to see per-target click tracking in action
- Complete the campaign when you're done — this disables the lure and phishlet on miraged
