"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";

export default function OpenPage() {
  const searchParams = useSearchParams();
  const repo = searchParams.get("repo") ?? "";
  const branch = searchParams.get("branch") ?? "";
  const [showInstall, setShowInstall] = useState(false);

  useEffect(() => {
    if (!repo || !branch) return;

    const deepLink = `chekout://open?repo=${encodeURIComponent(repo)}&branch=${encodeURIComponent(branch)}`;
    window.location.href = deepLink;

    const timer = setTimeout(() => {
      setShowInstall(true);
    }, 2500);

    return () => clearTimeout(timer);
  }, [repo, branch]);

  if (!repo || !branch) {
    return (
      <main style={styles.main}>
        <p style={styles.error}>Missing <code>repo</code> or <code>branch</code> query parameters.</p>
      </main>
    );
  }

  if (showInstall) {
    const installUrl = `https://chekout.sh?repo=${encodeURIComponent(repo)}&branch=${encodeURIComponent(branch)}`;
    return (
      <main style={styles.main}>
        <h1 style={styles.heading}>chekout.sh not installed</h1>
        <p style={styles.body}>
          Install the desktop app to open <strong>{repo}</strong> at branch <strong>{branch}</strong> in your local IDE.
        </p>
        <a href={installUrl} style={styles.button}>
          Install chekout.sh →
        </a>
      </main>
    );
  }

  return (
    <main style={styles.main}>
      <p style={styles.body}>Opening in your IDE…</p>
    </main>
  );
}

const styles = {
  main: {
    display: "flex",
    flexDirection: "column" as const,
    alignItems: "center",
    justifyContent: "center",
    minHeight: "100vh",
    padding: "2rem",
    fontFamily: "system-ui, sans-serif",
    gap: "1rem",
  },
  heading: {
    fontSize: "1.5rem",
    fontWeight: 600,
  },
  body: {
    fontSize: "1rem",
    color: "#555",
    textAlign: "center" as const,
  },
  error: {
    fontSize: "1rem",
    color: "#c00",
  },
  button: {
    marginTop: "0.5rem",
    padding: "0.75rem 1.5rem",
    background: "#171717",
    color: "#fff",
    borderRadius: "6px",
    fontSize: "0.95rem",
    textDecoration: "none",
  },
};
