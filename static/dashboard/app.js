(function () {
  const inputCidade = document.getElementById('cidade-input');
  const buscarBtn = document.getElementById('buscar-btn');
  const loading = document.getElementById('loading');
  const wsIndicator = document.getElementById('ws-indicator');
  const historicoList = document.getElementById('historico-list');
  const statUptime = document.getElementById('stat-uptime');
  const statHitRate = document.getElementById('stat-hit-rate');
  const statEntradas = document.getElementById('stat-entradas');

  const climaCidade = document.getElementById('clima-cidade');
  const climaUF = document.getElementById('clima-uf');
  const climaIcon = document.getElementById('clima-icon');
  const climaTemp = document.getElementById('clima-temp');
  const climaMin = document.getElementById('clima-min');
  const climaMax = document.getElementById('clima-max');
  const climaUmidade = document.getElementById('clima-umidade');
  const climaVento = document.getElementById('clima-vento');
  const climaCondicao = document.getElementById('clima-condicao');

  let wsAtual = null;
  let map = null;
  let markerAtual = null;
  let chart = null;
  let debounceTimer = null;
  let cidadeAtual = '';

  function toggleLoading(isLoading) {
    loading.classList.toggle('visible', Boolean(isLoading));
  }

  function getIconeCondicao(condicao) {
    const texto = String(condicao || '').toLowerCase();
    if (texto.includes('ensolarado')) return '☀️';
    if (texto.includes('nublado')) return '☁️';
    if (texto.includes('chuva') || texto.includes('garoa')) return '🌧️';
    if (texto.includes('parcialmente')) return '⛅';
    return '🌡️';
  }

  function formatarTemperatura(valor) {
    if (valor === undefined || valor === null || Number.isNaN(Number(valor))) {
      return '--°';
    }
    return `${Number(valor).toFixed(1).replace('.0', '')}°`;
  }

  function renderizarHistorico() {
    const historico = JSON.parse(localStorage.getItem('historico') || '[]');
    historicoList.innerHTML = '';

    historico.forEach((nome) => {
      const li = document.createElement('li');
      const button = document.createElement('button');
      button.type = 'button';
      button.textContent = nome;
      button.addEventListener('click', () => {
        inputCidade.value = nome;
        buscarCidade(nome);
      });
      li.appendChild(button);
      historicoList.appendChild(li);
    });
  }

  function salvarHistorico(nome) {
    const cidade = String(nome || '').trim();
    if (!cidade) return;

    const historicoAtual = JSON.parse(localStorage.getItem('historico') || '[]');
    const atualizado = [cidade, ...historicoAtual.filter((item) => item !== cidade)].slice(0, 5);
    localStorage.setItem('historico', JSON.stringify(atualizado));
    renderizarHistorico();
  }

  function atualizarIndicadorWS(conectado) {
    wsIndicator.textContent = conectado ? '🟢 Conectado' : '🔴 Desconectado';
  }

  function exibeClima(data) {
    if (!data) return;

    const clima = data.clima || data.dados || data;

    cidadeAtual = data.cidade || cidadeAtual;
    climaCidade.textContent = data.cidade || '--';
    climaUF.textContent = data.uf ? `UF: ${data.uf}` : 'UF: --';
    climaIcon.textContent = getIconeCondicao(clima.condicao);

    const tempAtual = clima.temperatura_atual ?? clima.tempAtual;
    const tempMin = clima.temperatura_min ?? clima.tempMin;
    const tempMax = clima.temperatura_max ?? clima.tempMax;
    const umidade = clima.umidade;
    const vento = clima.vento_kmh ?? clima.ventoKmh;
    const condicao = clima.condicao || '--';

    climaTemp.textContent = formatarTemperatura(tempAtual);
    climaMin.textContent = formatarTemperatura(tempMin);
    climaMax.textContent = formatarTemperatura(tempMax);
    climaUmidade.textContent = umidade !== undefined ? `${umidade}%` : '--%';
    climaVento.textContent = vento !== undefined ? `${Number(vento).toFixed(1).replace('.0', '')} km/h` : '-- km/h';
    climaCondicao.textContent = condicao;
  }

  function destruirChart() {
    if (chart) {
      chart.destroy();
      chart = null;
    }
  }

  function buscarPrevisao(nome) {
    if (!nome) return;

    fetch(`/api/v1/clima/${encodeURIComponent(nome)}/previsao`)
      .then((response) => {
        if (!response.ok) {
          throw new Error('Falha ao buscar previsão');
        }
        return response.json();
      })
      .then((dados) => {
        const labels = Array.isArray(dados) ? dados.map((item) => item.data) : [];
        const minValues = Array.isArray(dados) ? dados.map((item) => item.min) : [];
        const maxValues = Array.isArray(dados) ? dados.map((item) => item.max) : [];

        destruirChart();

        const ctx = document.getElementById('chart');
        chart = new Chart(ctx, {
          type: 'line',
          data: {
            labels,
            datasets: [
              {
                label: 'Mínima',
                data: minValues,
                borderColor: '#0f8b8d',
                backgroundColor: 'rgba(15, 139, 141, 0.12)',
                tension: 0.35,
                fill: true,
              },
              {
                label: 'Máxima',
                data: maxValues,
                borderColor: '#f59e0b',
                backgroundColor: 'rgba(245, 158, 11, 0.12)',
                tension: 0.35,
                fill: true,
              },
            ],
          },
          options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
              legend: { position: 'bottom' },
            },
            scales: {
              y: {
                ticks: {
                  callback(value) {
                    return `${value}°`;
                  },
                },
              },
            },
          },
        });
      })
      .catch(() => {
        destruirChart();
      });
  }

  function centralizaMapa(lat, lon, cidade) {
    if (!window.L) return;

    if (!map) {
      map = L.map('map').setView([lat, lon], 12);
      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; OpenStreetMap contributors',
      }).addTo(map);
    } else {
      map.setView([lat, lon], 12);
    }

    if (markerAtual) {
      map.removeLayer(markerAtual);
    }

    markerAtual = L.marker([lat, lon]).addTo(map);
    markerAtual.bindPopup(cidade).openPopup();
  }

  function conectarWS(nome) {
    if (wsAtual) {
      wsAtual.close();
      wsAtual = null;
    }

    if (!nome) {
      atualizarIndicadorWS(false);
      return;
    }

    wsAtual = new WebSocket(`ws://${window.location.host}/api/v1/ws/clima/${encodeURIComponent(nome)}`);

    wsAtual.onopen = () => atualizarIndicadorWS(true);
    wsAtual.onclose = () => atualizarIndicadorWS(false);
    wsAtual.onerror = () => atualizarIndicadorWS(false);
    wsAtual.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        if (payload && payload.evento === 'clima_atualizado') {
          exibeClima(payload.dados);
        }
      } catch (error) {
        void error;
      }
    };
  }

  function buscarCidade(nome) {
    const cidade = String(nome || '').trim();
    if (!cidade) {
      return;
    }

    toggleLoading(true);

    fetch(`/api/v1/clima/${encodeURIComponent(cidade)}`)
      .then((response) => {
        if (!response.ok) {
          throw new Error('Falha ao buscar clima');
        }
        return response.json();
      })
      .then((data) => {
        exibeClima(data);
        centralizaMapa(data.latitude, data.longitude, data.cidade);
        buscarPrevisao(cidade);
        conectarWS(cidade);
        salvarHistorico(cidade);
      })
      .catch(() => {
        atualizarIndicadorWS(false);
      })
      .finally(() => {
        toggleLoading(false);
      });
  }

  function atualizarStats() {
    fetch('/api/v1/health')
      .then((response) => response.json())
      .then((data) => {
        statUptime.textContent = data.uptime || '--';
        statHitRate.textContent = data.cache?.hit_rate || '--';
        statEntradas.textContent = data.cache?.entradas ?? '--';
        document.getElementById('stats-updated').textContent = 'Atualizado agora';
      })
      .catch(() => {
        document.getElementById('stats-updated').textContent = 'Indisponível';
      });
  }

  inputCidade.addEventListener('keyup', () => {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      buscarCidade(inputCidade.value.trim());
    }, 500);
  });

  buscarBtn.addEventListener('click', () => {
    buscarCidade(inputCidade.value.trim());
  });

  window.addEventListener('beforeunload', () => {
    if (wsAtual) {
      wsAtual.close();
    }
  });

  renderizarHistorico();
  atualizarIndicadorWS(false);
  atualizarStats();
  setInterval(atualizarStats, 30000);
})();
