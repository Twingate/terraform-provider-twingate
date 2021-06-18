const {danger, warn, fail} = require('danger')

const regexTitle = /(?<ticket>TG-\d\d\d\d) | .+/

const title = danger.github.pr.title.trim()
const body = danger.github.pr.body
const isUser = danger.github.pr.user.type === "User"

if (isUser) {
  // PR Title should match regexp
  if (!regexTitle.test(title)) {
    fail(`Please use standard PR title. (example: "TG-1234 | Some description")`)
  }

  // PR body has to contain the ticket from title
  const match = regexTitle.exec(title);
  if (match === null) {
    process.exit()
  }

  const ticket = match.groups.ticket;
  if (!body.includes(ticket)) {
    fail(`Please include the ticket (${ticket}) link in the PR body.`);
  }
}
