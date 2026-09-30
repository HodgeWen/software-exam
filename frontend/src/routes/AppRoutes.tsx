import { BrowserRouter, Navigate, Route, Routes } from "react-router";
import { LoginPage } from "../features/auth/LoginPage";
import { RegisterPage } from "../features/auth/RegisterPage";
import { ExamPage } from "../features/exam/ExamPage";
import { FavoritesPage } from "../features/favorites/FavoritesPage";
import { MistakeReviewPage } from "../features/mistakes/MistakeReviewPage";
import { MistakesPage } from "../features/mistakes/MistakesPage";
import { ChapterPracticePage } from "../features/practice/ChapterPracticePage";
import { RandomPracticePage } from "../features/practice/RandomPracticePage";
import { StatsPage } from "../features/stats/StatsPage";
import { SubjectSelectPage } from "./SubjectSelectPage";
import { SubjectPage } from "./SubjectPage";
import { RequireAuth } from "./RequireAuth";

// 路由树根组件：公开页仅登录/注册，其余一律经 RequireAuth 守卫
export function AppRoutes() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route
          path="/"
          element={
            <RequireAuth>
              <SubjectSelectPage />
            </RequireAuth>
          }
        />
        <Route
          path="/subjects/:subjectId"
          element={
            <RequireAuth>
              <SubjectPage />
            </RequireAuth>
          }
        />
        <Route
          path="/subjects/:subjectId/chapters/:chapterId"
          element={
            <RequireAuth>
              <ChapterPracticePage />
            </RequireAuth>
          }
        />
        <Route
          path="/subjects/:subjectId/random"
          element={
            <RequireAuth>
              <RandomPracticePage />
            </RequireAuth>
          }
        />
        <Route
          path="/papers/:paperId/exam"
          element={
            <RequireAuth>
              <ExamPage />
            </RequireAuth>
          }
        />
        <Route
          path="/mistakes"
          element={
            <RequireAuth>
              <MistakesPage />
            </RequireAuth>
          }
        />
        <Route
          path="/mistakes/review"
          element={
            <RequireAuth>
              <MistakeReviewPage />
            </RequireAuth>
          }
        />
        <Route
          path="/favorites"
          element={
            <RequireAuth>
              <FavoritesPage />
            </RequireAuth>
          }
        />
        <Route
          path="/stats"
          element={
            <RequireAuth>
              <StatsPage />
            </RequireAuth>
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
