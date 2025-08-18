import { SignJWT } from 'jose';

const myIpText = document.getElementById("myIpText")
const getMyIpBtn = document.getElementById("getMyIpBtn")
const copyMyIpBtn = document.getElementById("copyMyIpBtn")

getMyIpBtn.addEventListener("click", (e) => {
  getMyIpBtn.disabled = true
  copyMyIpBtn.disabled = true

  myIpText.textContent = "Loading..."

  fetch("/myip").then(res => res.text()).then(data => {
    myIpText.textContent = data
    copyMyIpBtn.disabled = false
  }).catch(e => {
    myIpText.textContent = "Error"
  }).finally(() => {
    getMyIpBtn.disabled = false
  })
})

document.getElementById("copyMyIpBtn").addEventListener("click", (e) => {
  navigator.clipboard.writeText(myIpText.textContent).then(() => {
    copyMyIpBtn.textContent = "Copied"
    setTimeout(() => copyMyIpBtn.textContent = "Copy My IP", 600)
  })
})


const signBtn = document.getElementById("signBtn")
const jwtSecretInput = document.getElementById("jwtSecretInput")
const domainPrefixInput = document.getElementById("domainPrefixInput")
const domainSelect = document.getElementById("domainSelect")
const signatureOutput = document.getElementById("signatureOutput")


signBtn.addEventListener("click", (e) => {
  const jwtSecret = jwtSecretInput.value
  const host = domainPrefixInput.value
  const domain = domainSelect.value

  if (!jwtSecret || !domain || !host) {
    alert("Please fill in all fields")
    return
  }

  const secret = new TextEncoder().encode(jwtSecret)
  new SignJWT({ host, domain })
    .setProtectedHeader({ alg: "HS256" })
    .sign(secret)
    .then(token => {
      signatureOutput.value = showToken(token)
    }).catch(e => {
      console.error(e)
      signatureOutput.value = "Error"
    })
})

const urlPrefix = location.protocol + "//" + location.host

const showToken = (token) => `Raw JWT Token:
${token}

Auto DDNS URL:
${urlPrefix}/ddns/v1/${token}

Manual DDNS URL:
${urlPrefix}/ddns/v1/${token}/<your-ip>
`
