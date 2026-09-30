import { useEffect, useState } from "react";

// 考试倒计时：以截止时间戳为准每秒刷新剩余秒数，到零停在下界不再回升。
// 归零只负责把 remaining 报出去，是否自动交卷由调用方决定
export function useExamCountdown(deadline: number): number {
  const [remaining, setRemaining] = useState(() => remainingSeconds(deadline));

  useEffect(() => {
    setRemaining(remainingSeconds(deadline));
    const timer = setInterval(() => setRemaining(remainingSeconds(deadline)), 1000);
    return () => clearInterval(timer);
  }, [deadline]);

  return remaining;
}

function remainingSeconds(deadline: number): number {
  return Math.max(0, Math.ceil((deadline - Date.now()) / 1000));
}
