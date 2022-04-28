const {danger, warn, fail} = require('danger')

const title = danger.github.pr.title.trim()
const body = danger.github.pr.body.trim()
const isUser = danger.github.pr.user.type === "User"

if (isUser) {
  if (body.includes("Please fill out as much as you can")) {
    fail(`Please include meaningful description.`);
  }

  if (body.includes("Make sure to document important changes")) {
    fail(`Please include meaningful description.`);
  }
}
