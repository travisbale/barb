# Barb Quickstart

This guide covers connecting Barb to the Mirage quickstart environment so you can run a full phishing campaign end-to-end — from email delivery through session capture — using the bundled target site.

## Prerequisites

Complete the [Mirage Quickstart](https://github.com/travisbale/mirage/blob/master/docs/quickstart.md) first. You should have:

- `miraged` and the target site running via Docker Compose
- `/etc/hosts` entries for `*.phish.local` and `*.target.local`
- The self-signed CA trusted in your browser

## 1. Build and start Barb

```bash
make build
./build/barb serve --debug
```

Open `http://localhost:8080` in your browser. On first login, Barb will prompt you to set a password.

## 2. Connect to miraged

Navigate to **Connections** and add a new miraged connection:

1. Find the enrollment token in the miraged logs:

   ```bash
   docker compose -f /path/to/mirage/examples/quickstart/docker-compose.yml logs miraged | grep "enroll with"
   ```

2. In Barb's connection form, enter:
   - **Address:** `127.0.0.1:443`
   - **Secret hostname:** `mgmt.phish.local`
   - **Token:** the token from the logs

Barb generates a keypair and enrolls with miraged automatically.

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

Navigate to **Phishlets** and create a new phishlet. Paste the contents of one of the example phishlet files from the mirage repo (e.g., `examples/phishlets/form-login.yaml`).

## 5. Create a target list

Navigate to **Target Lists** and create a list with at least one target. For local testing, any email address works — Mailpit accepts everything.

## 6. Create an email template

Navigate to **Templates** and create an email template. The template body supports Go template variables:

- `{{.FirstName}}` — target's first name
- `{{.LastName}}` — target's last name
- `{{.Email}}` — target's email address
- `{{.URL}}` — the unique lure URL (generated per-campaign)

Include `{{.URL}}` as a link in the body — this is what the target clicks to reach the phishing proxy.

## 7. Launch a campaign

Navigate to **Campaigns** and create a new campaign using the wizard:

1. Select the miraged connection
2. Select the phishlet and set the hostname to `login.phish.local`
3. Select the target list
4. Select the email template and SMTP profile
5. Set the redirect URL to `https://login.target.local:8443/demo-complete`
6. Start the campaign

Barb pushes the phishlet to miraged, creates a lure, and begins sending emails. Check Mailpit at `http://localhost:8025` to see the delivered emails, then click the lure URL to walk through the login flow.

As sessions are captured by miraged, Barb's session monitor picks them up in real time and correlates them to campaign targets.
