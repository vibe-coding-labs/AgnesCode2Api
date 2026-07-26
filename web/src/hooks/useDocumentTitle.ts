import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';

const TITLES: Record<string, string> = {
  '/': '数据概览 — AgnesCode 代理',
  '/accounts': '账号管理 — AgnesCode 代理',
  '/settings': '系统设置 — AgnesCode 代理',
};

const DEFAULT_TITLE = 'AgnesCode 代理';

const useDocumentTitle = () => {
  const location = useLocation();
  useEffect(() => {
    if (location.pathname.startsWith('/accounts/')) {
      const key = decodeURIComponent(location.pathname.replace('/accounts/', ''));
      document.title = `${key} — 账号详情 — AgnesCode 代理`;
    } else {
      document.title = TITLES[location.pathname] || DEFAULT_TITLE;
    }
  }, [location.pathname]);
};

export default useDocumentTitle;
