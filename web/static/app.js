(function () {
    "use strict";

    const state = { teams: {}, currentWeek: 1, totalWeeks: 6 };

    async function api(method, path, body) {
        const opts = { method, headers: {} };
        if (body) {
            opts.headers["Content-Type"] = "application/json";
            opts.body = JSON.stringify(body);
        }
        const res = await fetch(path, opts);
        const data = await res.json();
        if (!res.ok) {
            throw new Error(data.error || data.message || `HTTP ${res.status}`);
        }
        return data;
    }

    function showToast(msg) {
        const container = document.getElementById("toast-container");
        const el = document.createElement("div");
        el.className = "toast";
        el.textContent = msg;
        container.appendChild(el);
        setTimeout(() => el.remove(), 4000);
    }

    function buildTeamMap(matches) {
        // We need the table data to get team names mapped to IDs
    }

    function updateWeekLabel(week, total) {
        state.currentWeek = week;
        state.totalWeeks = total;
        const label = document.getElementById("week-label");
        if (week > total) {
            label.textContent = "SEASON COMPLETE";
            document.getElementById("btn-next-week").disabled = true;
            document.getElementById("btn-play-all").disabled = true;
        } else {
            label.textContent = `MATCHWEEK ${week} OF ${total}`;
            document.getElementById("btn-next-week").disabled = false;
            document.getElementById("btn-play-all").disabled = false;
        }
    }

    function renderStandings(rows, animate) {
        const tbody = document.getElementById("standings-body");
        const oldPositions = {};
        tbody.querySelectorAll("tr").forEach((tr) => {
            const id = tr.dataset.teamId;
            if (id) oldPositions[id] = tr.rowIndex;
        });

        tbody.innerHTML = "";
        rows.forEach((row, i) => {
            state.teams[row.team_id] = row.team_name;
            const tr = document.createElement("tr");
            tr.dataset.teamId = row.team_id;
            if (i === 0) tr.classList.add("leader");
            if (animate) {
                tr.classList.add("stagger-row");
                tr.style.animationDelay = `${i * 40}ms`;
            }
            if (!animate && oldPositions[row.team_id] !== undefined && oldPositions[row.team_id] !== i + 1) {
                tr.classList.add("highlight");
            }
            tr.innerHTML = `
                <td class="col-pos">${i + 1}</td>
                <td class="col-team">${row.team_name}</td>
                <td class="col-num">${row.played}</td>
                <td class="col-num">${row.wins}</td>
                <td class="col-num">${row.draws}</td>
                <td class="col-num">${row.losses}</td>
                <td class="col-num">${row.goals_for}</td>
                <td class="col-num">${row.goals_against}</td>
                <td class="col-num">${row.goal_difference}</td>
                <td class="col-num col-pts">${row.points}</td>
            `;
            tbody.appendChild(tr);
        });
    }

    function renderResults(matches, weekNum, animate) {
        const title = document.getElementById("results-title");
        title.textContent = weekNum ? `MATCHWEEK ${weekNum} RESULTS` : "LATEST RESULTS";

        const list = document.getElementById("results-list");
        list.innerHTML = "";

        if (!matches || matches.length === 0) {
            list.innerHTML = '<div style="color:var(--muted);font-size:12px;padding:8px;">No results yet</div>';
            return;
        }

        matches.forEach((m, i) => {
            const homeName = state.teams[m.home_team_id] || `Team ${m.home_team_id}`;
            const awayName = state.teams[m.away_team_id] || `Team ${m.away_team_id}`;
            const hg = m.home_goals !== null ? m.home_goals : "-";
            const ag = m.away_goals !== null ? m.away_goals : "-";

            const el = document.createElement("div");
            el.className = "result-item" + (animate ? " slide-in" : "");
            if (animate) el.style.animationDelay = `${i * 100}ms`;
            el.innerHTML = `
                <span class="home-team">${homeName}</span>
                <span class="score">${hg} - ${ag}</span>
                <span class="away-team">${awayName}</span>
            `;
            list.appendChild(el);
        });
    }

    function renderPredictions(predictions) {
        const list = document.getElementById("predictions-list");
        list.innerHTML = "";

        if (state.currentWeek <= 4 && state.currentWeek <= state.totalWeeks) {
            list.innerHTML = '<div class="predictions-locked">ODDS UNLOCK AFTER WEEK 4</div>';
            return;
        }

        if (!predictions || predictions.length === 0) {
            list.innerHTML = '<div class="predictions-locked">NO PREDICTIONS AVAILABLE</div>';
            return;
        }

        predictions.forEach((p) => {
            const pct = Math.round(p.probability * 100);
            const el = document.createElement("div");
            el.className = "prediction-item";
            el.innerHTML = `
                <span class="team-name">${p.team_name}</span>
                <div class="bar-track"><div class="bar-fill"></div></div>
                <span class="pct">${pct}%</span>
            `;
            list.appendChild(el);
            requestAnimationFrame(() => {
                el.querySelector(".bar-fill").style.width = `${pct}%`;
            });
        });
    }

    function renderAllMatches(matches) {
        const container = document.getElementById("all-matches");
        container.innerHTML = "";

        const byWeek = {};
        matches.forEach((m) => {
            if (!byWeek[m.week]) byWeek[m.week] = [];
            byWeek[m.week].push(m);
        });

        const weeks = Object.keys(byWeek).sort((a, b) => a - b);
        weeks.forEach((week) => {
            const group = document.createElement("div");
            group.className = "week-group";
            group.innerHTML = `<div class="week-group-title">WEEK ${week}</div>`;

            byWeek[week].forEach((m) => {
                const homeName = state.teams[m.home_team_id] || `Team ${m.home_team_id}`;
                const awayName = state.teams[m.away_team_id] || `Team ${m.away_team_id}`;
                const hg = m.home_goals !== null ? m.home_goals : "-";
                const ag = m.away_goals !== null ? m.away_goals : "-";

                const row = document.createElement("div");
                row.className = "match-row " + (m.played ? "played" : "unplayed");
                row.dataset.matchId = m.id;
                row.innerHTML = `
                    <span class="home-team">${homeName}</span>
                    <span class="score">${hg} - ${ag}</span>
                    <span class="away-team">${awayName}</span>
                `;

                if (m.played) {
                    row.addEventListener("click", () => openEditForm(row, m));
                }

                group.appendChild(row);
            });

            container.appendChild(group);
        });
    }

    function openEditForm(row, match) {
        const existing = document.querySelector(".edit-form");
        if (existing) existing.replaceWith(existing._originalRow);

        const homeName = state.teams[match.home_team_id] || `Team ${match.home_team_id}`;
        const awayName = state.teams[match.away_team_id] || `Team ${match.away_team_id}`;

        const form = document.createElement("div");
        form.className = "edit-form";
        form._originalRow = row;
        form.innerHTML = `
            <span class="team-label">${homeName}</span>
            <input type="number" min="0" max="20" value="${match.home_goals || 0}" id="edit-home">
            <span style="color:var(--muted)">-</span>
            <input type="number" min="0" max="20" value="${match.away_goals || 0}" id="edit-away">
            <span class="team-label">${awayName}</span>
            <span class="spacer"></span>
            <button class="btn btn-primary" id="edit-save">SAVE</button>
            <button class="btn" id="edit-cancel">CANCEL</button>
        `;

        row.replaceWith(form);

        form.querySelector("#edit-cancel").addEventListener("click", () => {
            form.replaceWith(row);
        });

        form.querySelector("#edit-save").addEventListener("click", async () => {
            const hg = parseInt(form.querySelector("#edit-home").value, 10);
            const ag = parseInt(form.querySelector("#edit-away").value, 10);
            try {
                await api("PUT", `/matches/${match.id}`, { home_goals: hg, away_goals: ag });
                form.replaceWith(row);
                await refreshAll(false);
            } catch (e) {
                showToast(e.message);
            }
        });
    }

    async function fetchPredictions() {
        if (state.currentWeek <= 4 && state.currentWeek <= state.totalWeeks) {
            renderPredictions(null);
            return;
        }
        try {
            const data = await api("GET", "/predictions");
            renderPredictions(data.predictions);
        } catch (e) {
            renderPredictions(null);
        }
    }

    async function refreshAll(animate) {
        try {
            const [weekData, tableData, matchData] = await Promise.all([
                api("GET", "/league/week"),
                api("GET", "/league/table"),
                api("GET", "/matches"),
            ]);

            updateWeekLabel(weekData.current_week, weekData.total_weeks);
            renderStandings(tableData.table || [], animate);

            const allMatches = matchData.matches || [];
            const latestWeek = weekData.current_week > weekData.total_weeks
                ? weekData.total_weeks
                : weekData.current_week - 1;
            const latestResults = allMatches.filter((m) => m.played && m.week === latestWeek);
            renderResults(latestResults, latestWeek > 0 ? latestWeek : null, animate);
            renderAllMatches(allMatches);

            const hasPlayed = allMatches.some((m) => m.played);
            const content = document.getElementById("all-matches");
            const arrow = document.getElementById("toggle-arrow");
            if (hasPlayed) {
                content.classList.remove("collapsed");
                arrow.classList.add("open");
            } else {
                content.classList.add("collapsed");
                arrow.classList.remove("open");
            }

            await fetchPredictions();
        } catch (e) {
            showToast(e.message);
        }
    }

    async function handleNextWeek() {
        try {
            const data = await api("POST", "/league/next-week");
            const weekMatches = data.matches || [];
            // Build temporary results from response (uses team names)
            const list = document.getElementById("results-list");
            list.innerHTML = "";
            const title = document.getElementById("results-title");
            title.textContent = `MATCHWEEK ${data.week} RESULTS`;

            weekMatches.forEach((m, i) => {
                const el = document.createElement("div");
                el.className = "result-item slide-in";
                el.style.animationDelay = `${i * 100}ms`;
                el.innerHTML = `
                    <span class="home-team">${m.home_team}</span>
                    <span class="score">${m.home_goals} - ${m.away_goals}</span>
                    <span class="away-team">${m.away_team}</span>
                `;
                list.appendChild(el);
            });

            await refreshAll(false);
        } catch (e) {
            showToast(e.message);
        }
    }

    async function handlePlayAll() {
        if (!confirm("Simulate remaining weeks?")) return;
        try {
            const data = await api("POST", "/league/play-all");
            const weeks = data.weeks || [];

            for (let i = 0; i < weeks.length; i++) {
                const wk = weeks[i];
                const list = document.getElementById("results-list");
                list.innerHTML = "";
                const title = document.getElementById("results-title");
                title.textContent = `MATCHWEEK ${wk.week} RESULTS`;

                wk.matches.forEach((m, j) => {
                    const el = document.createElement("div");
                    el.className = "result-item slide-in";
                    el.style.animationDelay = `${j * 100}ms`;
                    el.innerHTML = `
                        <span class="home-team">${m.home_team}</span>
                        <span class="score">${m.home_goals} - ${m.away_goals}</span>
                        <span class="away-team">${m.away_team}</span>
                    `;
                    list.appendChild(el);
                });

                await new Promise((r) => setTimeout(r, 600));
            }

            // Flash champion
            if (data.final_table && data.final_table.length > 0) {
                renderStandings(data.final_table, false);
                const tbody = document.getElementById("standings-body");
                const firstRow = tbody.querySelector("tr");
                if (firstRow) {
                    firstRow.classList.add("highlight");
                    setTimeout(() => firstRow.classList.remove("highlight"), 1500);
                }
            }

            await refreshAll(false);
        } catch (e) {
            showToast(e.message);
        }
    }

    async function handleReset() {
        if (!confirm("Reset the entire season?")) return;
        try {
            await api("POST", "/league/reset");
            await refreshAll(true);
        } catch (e) {
            showToast(e.message);
        }
    }

    function setupToggle() {
        const toggle = document.getElementById("toggle-all-matches");
        const content = document.getElementById("all-matches");
        const arrow = document.getElementById("toggle-arrow");
        toggle.addEventListener("click", () => {
            content.classList.toggle("collapsed");
            arrow.classList.toggle("open");
        });
    }

    document.addEventListener("DOMContentLoaded", () => {
        document.getElementById("btn-next-week").addEventListener("click", handleNextWeek);
        document.getElementById("btn-play-all").addEventListener("click", handlePlayAll);
        document.getElementById("btn-reset").addEventListener("click", handleReset);
        setupToggle();
        refreshAll(true);
    });
})();
