package constants

const TestConfigYAML = `
ident:
  server_name: test
  base_url: "http://127.0.0.1:9999"
  signing_key:
    algo: ed25519
    id: 0
    seed: ahphigh9jahchiequiechee4pha1Atuv
  invites:
    email_template:
      text: "/tmp/ident_invite_template_txt"
      html: "/tmp/ident_invite_template_html"
    subject_template: "{{.SenderDisplayName}} invited you to Matrix!"

http:
  listen_addr: "127.0.0.1:9999"

database:
  driver: sqlite3
  conn_string: ":memory:"

email:
  from: "Ident <ident@example.com>"
  smtp:
    hostname: mail.example.com
    port: 465
    username: "ident@example.com"
    password: somepassword
    enable_tls: on
`
