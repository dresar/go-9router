import { Suspense } from "react";
import { CardSkeleton } from "@/shared/components/Loading";
import PlaygroundClient from "./PlaygroundClient";

export const metadata = {
  title: "Playground | Go 9Router",
  description: "Test ChatGPT, Claude, Gemini, and Image Generation directly in your browser with auto-configured API keys.",
};

export default function PlaygroundPage() {
  return (
    <Suspense fallback={<CardSkeleton />}>
      <PlaygroundClient />
    </Suspense>
  );
}
