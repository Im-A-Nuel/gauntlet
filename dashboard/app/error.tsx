'use client';
export default function ErrorPage({reset}:{reset:()=>void}) {return <main className="loading-screen"><h1>The report could not be loaded.</h1><p>Check the run directory, then try again.</p><button onClick={reset}>Retry report</button></main>}
