// --- Global State ---
let accessToken = null;
let provider;
const API_BASE_URL = "http://localhost:8080";
const GANACHE_CHAIN_ID = "0x0539"; // 1337 in hex
const GANACHE_NETWORK = {
  chainId: GANACHE_CHAIN_ID,
  chainName: "Ganache Local",
  nativeCurrency: {
    name: "Ethereum",
    symbol: "ETH",
    decimals: 18,
  },
  rpcUrls: ["http://127.0.0.1:7545"],
};

// --- DOM Elements ---
const authSection = document.getElementById("auth-section");
const dashboardSection = document.getElementById("dashboard");
const signupForm = document.getElementById("signup-form");
const loginForm = document.getElementById("login-form");
const authResponseEl = document.getElementById("auth-response");
const networkStatusEl = document.getElementById("network-status");
const switchNetworkButton = document.getElementById("switch-network-button");

// --- API Helper ---
async function apiCall(
  endpoint,
  method = "GET",
  body = null,
  requiresAuth = true,
) {
  const headers = { "Content-Type": "application/json" };
  if (requiresAuth && accessToken) {
    headers["Authorization"] = "Bearer " + accessToken;
  }

  const config = {
    method,
    headers,
    credentials: "include", // Send cookies (for refresh token)
  };

  if (body) {
    config.body = JSON.stringify(body);
  }

  let response = await fetch(API_BASE_URL + endpoint, config);

  if (response.status === 401 && requiresAuth) {
    try {
      const refreshResponse = await fetch(API_BASE_URL + "/auth/refresh", {
        method: "POST",
        credentials: "include",
      });
      if (!refreshResponse.ok) {
        logout();
        throw new Error("Session expired. Please log in again.");
      }
      const data = await refreshResponse.json();
      accessToken = data.access_token;

      // Retry the original request with the new token
      headers["Authorization"] = "Bearer " + accessToken;
      response = await fetch(API_BASE_URL + endpoint, config);
    } catch (error) {
      return Promise.reject(error);
    }
  }

  return response;
}

// --- Authentication ---
signupForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  const formData = new FormData(signupForm);
  const data = Object.fromEntries(formData.entries());

  try {
    const response = await apiCall("/auth/signup", "POST", data, false);
    const result = await response.json();
    if (!response.ok) throw new Error(result.error || "Signup failed");

    authResponseEl.innerHTML = `<p class="success">Account created successfully! You can now log in.</p>`;
    signupForm.reset();
  } catch (error) {
    authResponseEl.innerHTML = `<p class="error">${error.message}</p>`;
  }
});

loginForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  const formData = new FormData(loginForm);
  const data = Object.fromEntries(formData.entries());

  try {
    const response = await apiCall("/auth/login", "POST", data, false);
    const result = await response.json();
    if (!response.ok) throw new Error(result.error || "Login failed");

    accessToken = result.access_token;
    authResponseEl.innerHTML = `<p class="success">Login successful!</p>`;
    setTimeout(() => {
      authSection.classList.add("hidden");
      dashboardSection.classList.remove("hidden");
      renderDashboard();
    }, 500);
  } catch (error) {
    authResponseEl.innerHTML = `<p class="error">${error.message}</p>`;
  }
});

async function logout() {
  try {
    await apiCall("/auth/logout", "POST");
  } catch (error) {
    console.error("Logout failed, clearing session anyway.", error);
  } finally {
    accessToken = null;
    dashboardSection.classList.add("hidden");
    dashboardSection.innerHTML = ""; // Clear dashboard content
    authSection.classList.remove("hidden");
    authResponseEl.innerHTML = `<p>You have been logged out.</p>`;
  }
}

// --- Dashboard Rendering and Logic ---
function renderDashboard() {
  const template = document.getElementById("dashboard-template");
  const dashboardContent = template.content.cloneNode(true);
  dashboardSection.appendChild(dashboardContent);

  // Attach event listeners to the new dashboard elements
  document.getElementById("logout-button").addEventListener("click", logout);
  document
    .getElementById("connect-wallet-button")
    .addEventListener("click", connectWallet);
  document
    .getElementById("refresh-balances-button")
    .addEventListener("click", fetchBalances);
  document
    .getElementById("refresh-history-button")
    .addEventListener("click", fetchPaymentHistory);
  document
    .getElementById("create-payment-form")
    .addEventListener("submit", sendPayment);
  document
    .getElementById("search-form")
    .addEventListener("submit", searchByPhone);

  // Initial data load
  fetchPaymentStats();
  fetchPaymentHistory();
  initEthers();
}

async function fetchBalances() {
  const walletInfoEl = document.getElementById("wallet-info");
  walletInfoEl.innerHTML = `<p aria-busy="true">Fetching balances...</p>`;
  try {
    const response = await apiCall("/wallet/balances");
    const result = await response.json();
    if (!response.ok) throw new Error(result.error);

    if (result.wallets && result.wallets.length > 0) {
      let html = "<ul>";
      result.wallets.forEach((w) => {
        html += `<li><strong>${w.address}</strong>: ${w.formatted}</li>`;
      });
      html += "</ul>";
      walletInfoEl.innerHTML = html;
    } else {
      walletInfoEl.innerHTML = `<p>No wallets connected yet.</p>`;
    }
  } catch (error) {
    walletInfoEl.innerHTML = `<p class="error">${error.message}</p>`;
  }
}

async function fetchPaymentStats() {
  const statsEl = document.getElementById("payment-stats");
  try {
    const response = await apiCall("/payments/stats");
    const stats = await response.json();
    if (!response.ok) throw new Error(stats.error);
    statsEl.innerHTML = `
                <p><strong>Total Payments:</strong> ${stats.total_payments}</p>
                <p><strong>Confirmed:</strong> ${stats.confirmed}</p>
                <p><strong>Pending:</strong> ${stats.pending}</p>
            `;
  } catch (error) {
    statsEl.innerHTML = `<p class="error">Could not load stats.</p>`;
  }
}

async function fetchPaymentHistory() {
  const historyEl = document.getElementById("payment-list");
  historyEl.innerHTML = `<p aria-busy="true">Fetching history...</p>`;
  try {
    const response = await apiCall("/payments");
    const result = await response.json();
    if (!response.ok) throw new Error(result.error);

    if (result.payments && result.payments.length > 0) {
      let html =
        "<table><thead><tr><th>To</th><th>Amount</th><th>Status</th><th>Date</th></tr></thead><tbody>";
      result.payments.forEach((p) => {
        html += `<tr>
                        <td><small>${p.to_address}</small></td>
                        <td>${ethers.formatEther(p.amount)} ETH</td>
                        <td>${p.status}</td>
                        <td>${new Date(p.created_at).toLocaleString()}</td>
                    </tr>`;
      });
      html += "</tbody></table>";
      historyEl.innerHTML = html;
    } else {
      historyEl.innerHTML = `<p>No payments found.</p>`;
    }
  } catch (error) {
    historyEl.innerHTML = `<p class="error">${error.message}</p>`;
  }
}

// --- Search and Payment Functions ---
async function searchByPhone(e) {
  e.preventDefault();
  const phone = document.getElementById("search-phone").value;
  const resultsEl = document.getElementById("search-results");
  resultsEl.innerHTML = `<p aria-busy="true">Searching...</p>`;

  try {
    const response = await apiCall(`/wallet/addresses/${phone}`);
    const addresses = await response.json();
    if (!response.ok) throw new Error(addresses.error || "Search failed");

    if (addresses && addresses.length > 0) {
      let html = "<p>Click an address to pay:</p><ul>";
      addresses.forEach((addr) => {
        html += `<li class="address-item" onclick="useAddress('${addr.address}')"><code>${addr.address}</code></li>`;
      });
      html += "</ul>";
      resultsEl.innerHTML = html;
    } else {
      resultsEl.innerHTML = `<p>No wallet addresses found for that phone number.</p>`;
    }
  } catch (error) {
    resultsEl.innerHTML = `<p class="error">${error.message}</p>`;
  }
}

function useAddress(address) {
  const toAddressEl = document.getElementById("to_address");
  toAddressEl.value = address;
  toAddressEl.scrollIntoView({ behavior: "smooth", block: "center" });
  toAddressEl.focus();
}

// --- MetaMask / Ethers.js Functions ---
async function initEthers() {
  if (typeof window.ethereum === "undefined") {
    networkStatusEl.innerHTML = "MetaMask not installed";
    networkStatusEl.className = "error";
    return false;
  }
  provider = new ethers.BrowserProvider(window.ethereum);

  const network = await provider.getNetwork();
  if (network.chainId !== BigInt(GANACHE_CHAIN_ID)) {
    networkStatusEl.innerHTML = `Incorrect network. Please switch to Ganache Local (Chain ID ${parseInt(
      GANACHE_CHAIN_ID,
      16,
    )}).`;
    networkStatusEl.className = "error";
    switchNetworkButton.classList.remove("hidden");
    return false;
  }

  networkStatusEl.innerHTML = `Connected to Ganache Local (Chain ID: ${network.chainId})`;
  networkStatusEl.className = "success";
  switchNetworkButton.classList.add("hidden");
  return true;
}

async function switchNetwork() {
  try {
    await window.ethereum.request({
      method: "wallet_switchEthereumChain",
      params: [{ chainId: GANACHE_CHAIN_ID }],
    });
  } catch (switchError) {
    // This error code indicates that the chain has not been added to MetaMask.
    if (switchError.code === 4902) {
      try {
        await window.ethereum.request({
          method: "wallet_addEthereumChain",
          params: [GANACHE_NETWORK],
        });
      } catch (addError) {
        console.error("Failed to add the Ganache network:", addError);
        alert("Failed to add the Ganache network. Please add it manually.");
      }
    } else {
      console.error("Failed to switch network:", switchError);
      alert("Failed to switch to the Ganache network.");
    }
  }
}

switchNetworkButton.addEventListener("click", switchNetwork);

async function connectWallet() {
  if (!(await initEthers())) return;
  const walletInfoEl = document.getElementById("wallet-info");
  walletInfoEl.innerHTML = `<p aria-busy="true">Please sign the message in MetaMask...</p>`;
  try {
    const signer = await provider.getSigner();
    const address = await signer.getAddress();
    const message = "Connect wallet";
    const signature = await signer.signMessage(message);

    const response = await apiCall("/wallet/connect", "POST", {
      message,
      signature,
    });
    const result = await response.json();

    if (!response.ok) throw new Error(result.error);

    walletInfoEl.innerHTML = `<p class="success">${result.message} <br>Address: ${result.walletAddress}</p>`;
    fetchBalances();
  } catch (err) {
    walletInfoEl.innerHTML = `<p class="error">Error: ${err.message}</p>`;
  }
}

async function sendPayment(e) {
  e.preventDefault();
  if (!(await initEthers())) return;

  const toAddress = document.getElementById("to_address").value;
  const amountEth = document.getElementById("amount").value;
  const description = document.getElementById("description").value;
  const paymentResponseEl = document.getElementById("payment-response");

  if (!ethers.isAddress(toAddress)) {
    paymentResponseEl.innerHTML = `<p class="error">Invalid recipient address.</p>`;
    return;
  }

  try {
    paymentResponseEl.innerHTML = `<p aria-busy="true">Please confirm transaction in MetaMask...</p>`;
    const signer = await provider.getSigner();
    const tx = await signer.sendTransaction({
      to: toAddress,
      value: ethers.parseEther(amountEth),
    });

    paymentResponseEl.innerHTML = `<p aria-busy="true">Waiting for confirmation...</p>`;
    await tx.wait();

    const response = await apiCall("/payments", "POST", {
      to_address: toAddress,
      amount: ethers.parseEther(amountEth).toString(),
      transaction_hash: tx.hash,
      description: description,
    });

    const result = await response.json();
    if (!response.ok) throw new Error(result.error);

    paymentResponseEl.innerHTML = `<p class="success">${result.message}</p>`;
    fetchPaymentHistory();
    fetchPaymentStats();
    document.getElementById("create-payment-form").reset();
  } catch (err) {
    paymentResponseEl.innerHTML = `<p class="error">Error: ${err.message}</p>`;
  }
}
