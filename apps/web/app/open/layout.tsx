import type { Metadata } from "next";
import { Suspense } from "react";

export const metadata: Metadata = {
  robots: "noindex",
};

export default function OpenLayout({ children }: { children: React.ReactNode }) {
  return <Suspense>{children}</Suspense>;
}
