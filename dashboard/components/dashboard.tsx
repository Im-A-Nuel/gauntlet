"use client";

import { useEffect, useRef, useState } from "react";
import type { CSSProperties, PointerEvent as ReactPointerEvent } from "react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Bar,
  BarChart,
  CartesianGrid,
  LabelList,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import {
  compare,
  formatScore,
  labels,
  mutantsOf,
  statuses,
  verdict,
  type LocatedMutant,
  type Run,
  type Status,
} from "@/lib/schema";
import type { RunSource } from "@/lib/runs";

const views = [
  ["overview", "Overview"],
  ["matrix", "Kill matrix"],
  ["survivors", "Survivors"],
  ["compare", "Compare runs"],
];
const stamp = (date: string) =>
  new Intl.DateTimeFormat("en-GB", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "UTC",
  }).format(new Date(date)) + " UTC";
const shortId = (run: Run) => run.runId;
const runOptionLabel = (run: Run) => {
  const recorded = new Intl.DateTimeFormat("en-GB", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
    timeZone: "UTC",
  }).format(new Date(run.createdAt));
  const trigger =
    run.trigger === "ci"
      ? "CI"
      : run.trigger === "strengthen"
        ? "Strengthened"
        : run.trigger === "hook"
          ? "Hook"
          : "Manual";
  return `${trigger} · ${formatScore(run.totals.trustScore)}% · ${recorded} UTC`;
};
const duration = (ms: number) =>
  ms < 60000
    ? `${(ms / 1000).toFixed(1)}s`
    : `${Math.floor(ms / 60000)}m ${Math.round((ms % 60000) / 1000)}s`;
const symbols: Record<Status, string> = {
  killed: "×",
  survived: "!",
  timeout: "T",
  noCoverage: "○",
  ignored: "–",
  compileError: "C",
  runtimeError: "E",
};

function Outcome({ run }: { run: Run }) {
  const result = verdict(run);
  return (
    <span className={`outcome ${result}`}>
      {result === "pass"
        ? "Policy met"
        : result === "fail"
          ? "Policy not met"
          : "Insufficient evidence"}
    </span>
  );
}
function ScoreBar({
  value,
  label,
  kind = "trust",
}: {
  value: number | null;
  label: string;
  kind?: string;
}) {
  return (
    <div className="score-bar">
      <div>
        <span>{label}</span>
        <strong>
          {formatScore(value)}
          {value === null ? "" : "%"}
        </strong>
      </div>
      <div className="track">
        <span className={kind} style={{ width: `${value ?? 0}%` }} />
      </div>
    </div>
  );
}
function MutationStrip({
  mutants,
  onSelect,
}: {
  mutants: LocatedMutant[];
  onSelect: (m: LocatedMutant) => void;
}) {
  return (
    <div className="mutation-strip">
      {mutants.map((m) => (
        <button
          key={`${m.path}:${m.id}`}
          className={`mutation-cell ${m.status}`}
          onClick={() => onSelect(m)}
          title={`${labels[m.status]} · ${m.mutator} · line ${m.line}`}
          aria-label={`${labels[m.status]} mutant ${m.id} in ${m.path} line ${m.line}`}
        >
          {symbols[m.status]}
        </button>
      ))}
    </div>
  );
}
function Legend() {
  return (
    <div className="legend">
      {(["killed", "survived", "timeout", "noCoverage"] as Status[]).map(
        (s) => (
          <span key={s}>
            <i className={s}>{symbols[s]}</i>
            {labels[s]}
          </span>
        ),
      )}
    </div>
  );
}
function Diff({ mutant }: { mutant: LocatedMutant }) {
  return (
    <div className="diff">
      <div className="diff-line original">
        <span>Original</span>
        <pre>
          <code>{mutant.original || "(empty source)"}</code>
        </pre>
      </div>
      <div className="diff-line mutated">
        <span>Mutation</span>
        <pre>
          <code>{mutant.mutated || "(code removed)"}</code>
        </pre>
      </div>
    </div>
  );
}
function Inspector({
  mutant,
  onClose,
}: {
  mutant: LocatedMutant | null;
  onClose: () => void;
}) {
  const ref = useRef<HTMLDialogElement>(null);
  useEffect(() => {
    if (mutant && !ref.current?.open) ref.current?.showModal();
    else if (!mutant && ref.current?.open) ref.current?.close();
  }, [mutant]);
  return (
    <dialog
      ref={ref}
      className="inspector"
      onCancel={onClose}
      onClose={onClose}
      aria-labelledby="inspector-title"
    >
      {mutant && (
        <>
          <div className="section-head">
            <span className={`status-text ${mutant.status}`}>
              {labels[mutant.status]}
            </span>
            <button
              aria-label="Close mutant inspection"
              onClick={onClose}
              className="close-button"
            >
              ×
            </button>
          </div>
          <p className="eyebrow">
            Mutant {mutant.id} / line {mutant.line}
          </p>
          <h2 id="inspector-title">{mutant.mutator}</h2>
          <p className="mono wrap">{mutant.path}</p>
          <Diff mutant={mutant} />
          <p className="inspector-note">
            {mutant.status === "survived"
              ? "The test suite passed with this change. Add an assertion for the affected behavior, then run Gauntlet again."
              : mutant.status === "noCoverage"
                ? "No test exercised this mutation. Add coverage for the affected behavior before interpreting its score."
                : mutant.status === "timeout"
                  ? "The mutant exceeded the test time limit. Timeouts count as detected, but should be reviewed for slow or flaky tests."
                  : "This outcome was recorded by the mutation engine. Inspect the original behavior before changing tests."}
          </p>
        </>
      )}
    </dialog>
  );
}

function MutationCore({ run }: { run: Run }) {
  const score = run.totals.trustScore ?? 0;
  const detected = run.totals.killed + run.totals.timeout;
  const scored = detected + run.totals.survived;
  const coreStyle = {
    "--score-offset": 100 - score,
  } as CSSProperties;

  function positionCore(event: ReactPointerEvent<HTMLDivElement>) {
    if (
      event.pointerType === "touch" ||
      window.matchMedia("(prefers-reduced-motion: reduce)").matches
    )
      return;
    const bounds = event.currentTarget.getBoundingClientRect();
    const x = (event.clientX - bounds.left) / bounds.width - 0.5;
    const y = (event.clientY - bounds.top) / bounds.height - 0.5;
    event.currentTarget.style.setProperty("--core-rotate-y", `${x * 7}deg`);
    event.currentTarget.style.setProperty("--core-rotate-x", `${y * -5}deg`);
    event.currentTarget.style.setProperty("--core-light-x", `${50 + x * 24}%`);
    event.currentTarget.style.setProperty("--core-light-y", `${48 + y * 20}%`);
  }

  function resetCore(event: ReactPointerEvent<HTMLDivElement>) {
    event.currentTarget.style.setProperty("--core-rotate-y", "0deg");
    event.currentTarget.style.setProperty("--core-rotate-x", "0deg");
    event.currentTarget.style.setProperty("--core-light-x", "50%");
    event.currentTarget.style.setProperty("--core-light-y", "48%");
  }

  return (
    <div
      className="core-visual"
      style={coreStyle}
      onPointerMove={positionCore}
      onPointerLeave={resetCore}
      aria-hidden="true"
    >
      <div className="core-reticle core-reticle-outer" />
      <div className="core-reticle core-reticle-inner" />
      <svg className="score-orbit" viewBox="0 0 120 120">
        <circle
          className="score-orbit-track"
          cx="60"
          cy="60"
          r="57"
          pathLength="100"
        />
        <circle
          key={`${run.runId}-${score}`}
          className="score-orbit-progress"
          cx="60"
          cy="60"
          r="57"
          pathLength="100"
        />
      </svg>
      <div className="core-image">
        <Image
          src="/visuals/mutation-core.webp"
          alt=""
          fill
          preload
          sizes="(max-width: 760px) 90vw, (max-width: 1100px) 50vw, 520px"
        />
      </div>
      <div className="core-reading core-reading-detected">
        <span>Detected</span>
        <strong>{detected}</strong>
        <small>of {scored} scored</small>
      </div>
      <div className="core-reading core-reading-revision">
        <span>Revision</span>
        <strong>{run.headSha.slice(0, 7)}</strong>
        <small>{run.changedFiles.length} changed files</small>
      </div>
      <span className="core-axis core-axis-x" />
      <span className="core-axis core-axis-y" />
    </div>
  );
}

export function Dashboard({
  runs,
  source,
  view,
  selectedId,
  initialError,
}: {
  runs: Run[];
  source: RunSource;
  view: string;
  selectedId?: string;
  initialError?: string;
}) {
  const router = useRouter();
  const run = runs.find((r) => r.runId === selectedId) ?? runs[0];
  const [inspecting, setInspecting] = useState<LocatedMutant | null>(null);
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState("all");
  const [copied, setCopied] = useState(false);
  const [copyError, setCopyError] = useState("");
  const [refreshing, setRefreshing] = useState(false);
  const [beforeId, setBeforeId] = useState(
    runs[1]?.runId ?? runs[0]?.runId ?? "",
  );
  const [afterId, setAfterId] = useState(runs[0]?.runId ?? "");
  const [aId, bId] = [
    runs.find((r) => r.runId === beforeId) ?? runs[1] ?? runs[0],
    runs.find((r) => r.runId === afterId) ?? runs[0],
  ];
  const all = run ? mutantsOf(run) : [];
  const survivors = all.filter((m) => m.status === "survived");
  const filtered = all.filter(
    (m) =>
      (status === "all" || m.status === status) &&
      `${m.path} ${m.mutator} ${m.id}`
        .toLowerCase()
        .includes(query.toLowerCase()),
  );
  const survivorList = survivors
    .filter((m) =>
      `${m.path} ${m.mutator} ${m.id}`
        .toLowerCase()
        .includes(query.toLowerCase()),
    )
    .sort((a, b) => a.path.localeCompare(b.path) || a.line - b.line);
  useEffect(() => {
    setRefreshing(false);
  }, [runs, initialError]);
  async function copyCommand() {
    try {
      await navigator.clipboard.writeText("gauntlet strengthen --prepare-only");
      setCopied(true);
      setCopyError("");
    } catch {
      setCopyError("Copy unavailable. Select the command below.");
    }
  }
  const routeFor = (v: string) =>
    `${v === "overview" ? "/" : `/${v}`}${run ? `?run=${encodeURIComponent(run.runId)}` : ""}`;

  return (
    <>
      <a className="skip-link" href="#report">
        Skip to report
      </a>
      <header className="app-header">
        <Link className="wordmark" href="/" aria-label="Gauntlet overview">
          gauntlet<span> / </span>
        </Link>
        <span className="header-purpose">Adversarial code verification</span>
        <span className="source-label">
          {source === "sample" ? "Recorded demo" : "Local artifacts"}
        </span>
      </header>
      <div className="workspace-bar">
        <div>
          <span className="folder-mark" aria-hidden="true">
            ⌑
          </span>
          <strong>
            {source === "sample"
              ? "Pricing / cart service"
              : "Workspace report"}
          </strong>
          <span className="workspace-sub">Mutation testing</span>
        </div>
        <button
          className="quiet-button"
          onClick={() => {
            setRefreshing(true);
            router.refresh();
          }}
          disabled={refreshing}
        >
          {refreshing ? "Reading reports…" : "Refresh reports"}
        </button>
      </div>
      <nav className="main-nav" aria-label="Report views">
        {views.map(([id, label]) => (
          <Link
            key={id}
            href={routeFor(id)}
            aria-current={view === id ? "page" : undefined}
          >
            {label}
            {id === "survivors" && run && (
              <span className="count">{survivors.length}</span>
            )}
          </Link>
        ))}
      </nav>
      <main id="report" className="report" tabIndex={-1}>
        {initialError ? (
          <section className="empty-state" role="alert">
            <p className="eyebrow">Artifact read failed</p>
            <h1>We could not open this report.</h1>
            <p>{initialError}</p>
            <button onClick={() => router.refresh()}>
              Retry loading reports
            </button>
          </section>
        ) : !run ? (
          <section className="empty-state">
            <p className="eyebrow">No mutation runs yet</p>
            <h1>Give your tests a harder test.</h1>
            <p>
              Run Gauntlet in your project, then point GAUNTLET_RUNS_DIR at its
              .gauntlet/runs folder.
            </p>
            <code>
              gauntlet init
              <br />
              gauntlet run --changed
            </code>
            <button onClick={() => router.refresh()}>Check for reports</button>
          </section>
        ) : (
          <>
            {selectedId && selectedId !== run.runId && (
              <p role="status" className="notice">
                That run is unavailable. Showing the most recent report.
              </p>
            )}
            <div className="report-context">
              <span className="eyebrow">
                Verification report <span className="context-slash">/</span>{" "}
                {view === "overview"
                  ? "Overview"
                  : views.find((v) => v[0] === view)?.[1]}
              </span>
              <label className="run-picker">
                Run
                <select
                  value={run.runId}
                  onChange={(e) =>
                    router.push(
                      `${view === "overview" ? "/" : `/${view}`}?run=${encodeURIComponent(e.target.value)}`,
                    )
                  }
                >
                  {runs.map((r) => (
                    <option key={r.runId} value={r.runId} title={r.runId}>
                      {runOptionLabel(r)}
                    </option>
                  ))}
                </select>
              </label>
            </div>
            {view === "overview" && (
              <>
                <section
                  className={`hero-report hero-${verdict(run)}`}
                  aria-labelledby="hero-title"
                >
                  <div className="hero-atmosphere" aria-hidden="true" />
                  <div className="hero-main">
                    <div className="hero-kicker">
                      <span className="eyebrow">Adversarial verification</span>
                      <span className="hero-run-id mono">
                        Run {shortId(run)}
                      </span>
                    </div>
                    <h1 id="hero-title">
                      {verdict(run) === "pass"
                        ? "Evidence,\nunder pressure."
                        : "Your tests passed.\nThe mutations did too."}
                    </h1>
                    <p className="hero-lede">
                      Gauntlet challenged the changed code and measured which
                      faults the current suite could actually detect.
                    </p>
                    <div className="hero-score">
                      <span>{formatScore(run.totals.trustScore)}</span>
                      {run.totals.trustScore !== null && (
                        <span className="score-unit">%</span>
                      )}
                      <div className="score-description">
                        Trust Score<small>Mutation detection rate</small>
                      </div>
                    </div>
                    <p className="hero-footnote">
                      {run.totals.killed + run.totals.timeout} detected /{" "}
                      {run.totals.killed +
                        run.totals.timeout +
                        run.totals.survived}{" "}
                      scored mutants <span>·</span> {run.totals.noCoverage}{" "}
                      without coverage
                    </p>
                    <div className="mobile-policy-summary">
                      <Outcome run={run} />
                      <span>
                        {verdict(run) === "pass"
                          ? `Clears the ${run.threshold}% merge threshold`
                          : verdict(run) === "fail"
                            ? `Below the ${run.threshold}% merge threshold`
                            : "No scored mutations to evaluate"}
                      </span>
                    </div>
                  </div>
                  <MutationCore run={run} />
                  <aside className="verdict-panel">
                    <div className="section-head">
                      <span className="eyebrow">Merge policy</span>
                      <Outcome run={run} />
                    </div>
                    <h2>
                      {verdict(run) === "pass"
                        ? "Evidence clears the bar."
                        : verdict(run) === "unscored"
                          ? "There is no score to trust."
                          : `${survivors.length} mutations went unnoticed.`}
                    </h2>
                    <p>
                      {verdict(run) === "pass"
                        ? "The recorded score meets this run’s configured threshold. Review remaining survivors before merging."
                        : "Inspect the surviving changes and strengthen the assertions that should catch them."}
                    </p>
                    <ScoreBar
                      value={run.totals.lineCoverage}
                      label="Line coverage"
                      kind="coverage"
                    />
                    <ScoreBar
                      value={run.totals.trustScore}
                      label="Trust Score"
                    />
                    <div className="policy-threshold">
                      <span>Required Trust Score</span>
                      <strong>{run.threshold}%</strong>
                    </div>
                    <p className="fine-print">
                      Coverage measures execution. Mutation testing measures
                      whether tests detect injected changes. A passing score is
                      evidence, not proof of correctness.
                    </p>
                  </aside>
                </section>
                <dl className="run-facts">
                  <div>
                    <dt>Source revision</dt>
                    <dd className="mono">{run.headSha.slice(0, 10)}</dd>
                  </div>
                  <div>
                    <dt>Changed files</dt>
                    <dd>{run.changedFiles.length} source files</dd>
                  </div>
                  <div>
                    <dt>Run duration</dt>
                    <dd>{duration(run.totals.durationMs)}</dd>
                  </div>
                  <div>
                    <dt>Recorded at</dt>
                    <dd>{stamp(run.createdAt)}</dd>
                  </div>
                </dl>
                <section className="file-section">
                  <div className="section-head">
                    <div>
                      <p className="eyebrow">Where the gaps are</p>
                      <h2>Verification by file</h2>
                    </div>
                    <Link className="text-link" href={routeFor("matrix")}>
                      Inspect kill matrix <span aria-hidden="true">↗</span>
                    </Link>
                  </div>
                  <FileTable run={run} onSelect={setInspecting} />
                </section>
                <section className="next-step">
                  <div>
                    <p className="eyebrow">Close the loop</p>
                    <h2>Turn survivors into stronger tests.</h2>
                    <p>
                      Prepare a handoff for IBM Bob with the exact mutations the
                      tests missed.
                    </p>
                  </div>
                  <div className="command-block">
                    <code>gauntlet strengthen --prepare-only</code>
                    <button onClick={copyCommand}>
                      {copied ? "Command copied" : "Copy command"}
                    </button>
                    {copyError && <p role="status">{copyError}</p>}
                  </div>
                </section>
              </>
            )}
            {view === "matrix" && (
              <>
                <div className="view-heading">
                  <h1>Every mutation. Every outcome.</h1>
                  <p>
                    Read the test suite one injected change at a time. Select a
                    cell to inspect its diff.
                  </p>
                </div>
                <div className="section-head">
                  <h2>
                    {all.length} mutations across {run.files.length} files
                  </h2>
                  <Legend />
                </div>
                <label className="search-field">
                  Filter files
                  <input
                    placeholder="Search a source file…"
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                  />
                </label>
                <div className="matrix-list">
                  {run.files
                    .filter((f) =>
                      f.path.toLowerCase().includes(query.toLowerCase()),
                    )
                    .map((file) => (
                      <section key={file.path} className="matrix-file">
                        <div className="section-head">
                          <h3 className="mono">{file.path}</h3>
                          <span>{formatScore(file.trustScore)}% trust</span>
                        </div>
                        <MutationStrip
                          mutants={file.mutants.map((m) => ({
                            ...m,
                            path: file.path,
                          }))}
                          onSelect={setInspecting}
                        />
                        <p>
                          {
                            file.mutants.filter((m) => m.status === "survived")
                              .length
                          }{" "}
                          survived · {file.mutants.length} total
                        </p>
                      </section>
                    ))}
                </div>
                {!run.files.some((f) =>
                  f.path.toLowerCase().includes(query.toLowerCase()),
                ) && (
                  <p className="empty-inline">
                    No files match “{query}”. Clear the filter to see all files.
                  </p>
                )}
              </>
            )}
            {view === "survivors" && (
              <>
                <div className="view-heading">
                  <p className="eyebrow">Uncaught changes</p>
                  <h1>
                    {survivors.length === 1
                      ? "1 behavior gap remains."
                      : `${survivors.length} behavior gaps remain.`}
                  </h1>
                  <p>
                    Ordered by source location. Open a mutation to see the exact
                    behavior the current assertions missed.
                  </p>
                </div>
                <div className="filter-bar">
                  <label className="search-field">
                    Search mutations
                    <input
                      placeholder="File, operator, or mutant ID…"
                      value={query}
                      onChange={(e) => setQuery(e.target.value)}
                    />
                  </label>
                  <label>
                    Outcome
                    <select
                      value={status}
                      onChange={(e) => setStatus(e.target.value)}
                    >
                      <option value="all">Survivors only</option>
                      {statuses.map((s) => (
                        <option key={s} value={s}>
                          {labels[s]}
                        </option>
                      ))}
                    </select>
                  </label>
                </div>
                <div className="survivor-list">
                  {(status === "all" ? survivorList : filtered).map((m) => (
                    <button
                      className="survivor-row"
                      key={`${m.path}:${m.id}`}
                      onClick={() => setInspecting(m)}
                    >
                      <span className={`status-mark ${m.status}`}>
                        {symbols[m.status]}
                      </span>
                      <span>
                        <strong>{m.mutator}</strong>
                        <small className="mono">
                          {m.path}:{m.line}
                        </small>
                      </span>
                      <code>{m.mutated || "(code removed)"}</code>
                      <span className={`status-text ${m.status}`}>
                        {labels[m.status]}
                      </span>
                      <span aria-hidden="true">↗</span>
                    </button>
                  ))}
                </div>
                {!(status === "all" ? survivorList : filtered).length && (
                  <div className="empty-inline">
                    <h2>
                      {query
                        ? "No mutations match this search."
                        : survivors.length === 0 && status === "all"
                          ? "No surviving mutations."
                          : "No mutations have this outcome."}
                    </h2>
                    <p>
                      {query
                        ? "Try a file name or clear the filter."
                        : "Review the matrix for timeouts and changes without coverage."}
                    </p>
                  </div>
                )}
              </>
            )}
            {view === "compare" && aId && bId && (
              <>
                <div className="view-heading">
                  <p className="eyebrow">Before / after</p>
                  <h1>Did the tests get stronger?</h1>
                  <p>
                    Compare recorded evidence. Changes in file scope can change
                    the meaning of the score.
                  </p>
                </div>
                <div className="comparison-selectors">
                  <label>
                    Baseline run
                    <select
                      value={aId.runId}
                      onChange={(e) => setBeforeId(e.target.value)}
                    >
                      {runs.map((r) => (
                        <option key={r.runId} value={r.runId}>
                          {runOptionLabel(r)}
                        </option>
                      ))}
                    </select>
                  </label>
                  <span aria-hidden="true">→</span>
                  <label>
                    Follow-up run
                    <select
                      value={bId.runId}
                      onChange={(e) => setAfterId(e.target.value)}
                    >
                      {runs.map((r) => (
                        <option key={r.runId} value={r.runId}>
                          {runOptionLabel(r)}
                        </option>
                      ))}
                    </select>
                  </label>
                </div>
                <Comparison a={aId} b={bId} />
              </>
            )}
          </>
        )}
      </main>
      <footer className="app-footer">
        <span>
          gauntlet <span className="footer-divider">/</span> Test the tests.
        </span>
        <span>
          {source === "sample"
            ? "Recorded Stryker results · no live agent session"
            : "Local JSON artifacts · read-only dashboard"}
        </span>
      </footer>
      <Inspector mutant={inspecting} onClose={() => setInspecting(null)} />
    </>
  );
}

function FileTable({
  run,
  onSelect,
}: {
  run: Run;
  onSelect: (m: LocatedMutant) => void;
}) {
  return (
    <div
      className="file-table"
      role="table"
      aria-label="Mutation results by source file"
    >
      <div className="file-table-heading" role="row">
        <span role="columnheader">Source file</span>
        <span role="columnheader">Detected / scored</span>
        <span role="columnheader">Survived</span>
        <span role="columnheader">Trust Score</span>
      </div>
      {[...run.files]
        .sort((a, b) => (a.trustScore ?? -1) - (b.trustScore ?? -1))
        .map((file) => {
          const survived = file.mutants.filter((m) => m.status === "survived");
          const detected = file.mutants.filter(
            (m) => m.status === "killed" || m.status === "timeout",
          ).length;
          return (
            <div key={file.path} className="file-table-row" role="row">
              <span role="cell">
                <button
                  className="file-link mono"
                  onClick={() => {
                    const m = survived[0] ?? file.mutants[0];
                    if (m) onSelect({ ...m, path: file.path });
                  }}
                  disabled={!file.mutants.length}
                >
                  {file.path}
                </button>
                <small>{file.mutants.length} mutations</small>
              </span>
              <span role="cell" className="mono">
                {detected}{" "}
                <span className="muted">/ {detected + survived.length}</span>
              </span>
              <span
                role="cell"
                className={survived.length ? "survivor-count" : "muted"}
              >
                {survived.length}
              </span>
              <span role="cell" className="file-score">
                <span className="mini-track">
                  <i
                    style={{ width: `${file.trustScore ?? 0}%` }}
                    className={
                      (file.trustScore ?? 0) >= run.threshold ? "good" : "bad"
                    }
                  />
                </span>
                <strong>
                  {formatScore(file.trustScore)}
                  <small>%</small>
                </strong>
              </span>
            </div>
          );
        })}
      {!run.files.length && (
        <p className="empty-inline">
          No source files were included in this run.
        </p>
      )}
    </div>
  );
}
function Comparison({ a, b }: { a: Run; b: Run }) {
  const comparison = compare(a, b);
  const delta = comparison.delta.trustScore;
  const series = [
    { name: "Baseline", score: a.totals.trustScore },
    { name: "Follow-up", score: b.totals.trustScore },
  ];
  return (
    <>
      {a.runId === b.runId && (
        <p className="notice">
          The same run is selected twice. Choose another run to measure a
          change.
        </p>
      )}
      {!comparison.comparable && (
        <p className="notice">
          These runs have different source revisions or file scopes. The score
          difference is descriptive, not a controlled test-strength comparison.
        </p>
      )}
      <section className="comparison-summary">
        <div>
          <p className="eyebrow">Trust Score change</p>
          <div className="delta-number">
            {delta === null
              ? "N/A"
              : `${delta > 0 ? "+" : ""}${delta.toFixed(1)}`}
            <span>pts</span>
          </div>
          <div className="survivor-shift">
            <span>Surviving mutations</span>
            <strong>
              {a.totals.survived} <span aria-hidden="true">→</span>{" "}
              {b.totals.survived}
            </strong>
          </div>
          <Outcome run={b} />
        </div>
        <div
          className="comparison-chart"
          role="img"
          aria-label={`Baseline Trust Score ${formatScore(a.totals.trustScore)} percent; follow-up ${formatScore(b.totals.trustScore)} percent`}
        >
          <ResponsiveContainer width="100%" height={220}>
            <BarChart
              data={series}
              layout="vertical"
              margin={{ left: 12, right: 52 }}
            >
              <CartesianGrid stroke="var(--border)" horizontal={false} />
              <XAxis
                type="number"
                domain={[0, 100]}
                stroke="var(--muted)"
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                type="category"
                dataKey="name"
                stroke="var(--muted)"
                tickLine={false}
                axisLine={false}
                width={80}
              />
              <Tooltip
                cursor={false}
                contentStyle={{
                  background: "var(--panel)",
                  border: "1px solid var(--border)",
                  color: "var(--text)",
                }}
              />
              <Bar
                dataKey="score"
                name="Trust Score"
                fill="var(--accent)"
                barSize={28}
                isAnimationActive={false}
              >
                <LabelList
                  dataKey="score"
                  position="right"
                  fill="var(--text)"
                  fontSize={12}
                  formatter={(value) => `${value}%`}
                />
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      </section>
      <div className="section-head">
        <h2>What changed by file</h2>
        <span className="muted">Baseline → Follow-up</span>
      </div>
      <div className="comparison-files">
        {comparison.perFile.map((f) => (
          <div key={f.path}>
            <span className="mono">{f.path}</span>
            <span>
              {formatScore(f.a)}% <span className="muted">→</span>{" "}
              <strong>{formatScore(f.b)}%</strong>
            </span>
          </div>
        ))}
      </div>
      <p className="fine-print">
        Both results are recorded artifacts; this view does not invoke an AI
        agent.
      </p>
    </>
  );
}
