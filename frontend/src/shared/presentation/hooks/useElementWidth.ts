"use client";

import { useEffect, useRef, useState } from "react";

export function useElementWidth(fallback = 0): [React.RefObject<HTMLDivElement>, number] {
  const ref = useRef<HTMLDivElement>(null!);
  const [width, setWidth] = useState(fallback);

  useEffect(() => {
    if (!ref.current) return;
    const observer = new ResizeObserver(([entry]) => {
      setWidth(entry.contentRect.width);
    });
    observer.observe(ref.current);
    setWidth(ref.current.getBoundingClientRect().width);
    return () => observer.disconnect();
  }, []);

  return [ref, width];
}
