"use client";

import { useState } from "react";

type NavItem = {
  icon: string;
  label: string;
  link: string;
  children?: string[];
};

const NAV_PRIMARY: NavItem[] = [
  { icon: "dashboard", label: "Dashboard", link: "/" },
  { icon: "auto_stories", label: "Ediciones", link: "/ediciones" },
  { icon: "notifications", label: "Notificaciones", link: "/notificaciones" },
  { icon: "settings", label: "Configuración", link: "/configuracion" },
];

const NAV_SECONDARY: NavItem[] = [
  { icon: "help", label: "Soporte", link: "/soporte" },
  { icon: "logout", label: "Cerrar sesión", link: "/cerrar-sesion" },
];

export function Sidebar() {
  const [collapsed, setCollapsed] = useState(false);
  const [openDropdown, setOpenDropdown] = useState<string | null>(null);

  const toggleDropdown = (label: string) => {
    setOpenDropdown((prev) => (prev === label ? null : label));
  };

  return (
    <>
      {/* Mobile button */}
      <button
        className="sidebar-menu-button"
        onClick={() => setCollapsed((v) => !v)}
      >
        <span className="material-symbols-rounded">menu</span>
      </button>

      <aside className={`sidebar${collapsed ? " collapsed" : ""}`}>
        <header className="sidebar-header">
          <a href="/" className="header-logo">
            <span style={{ fontSize: "1.6rem", color: "var(--sidebar-text)", fontWeight: 700, letterSpacing: "-1px" }}>{collapsed ? "FD" : "FlowDesign"}</span>
          </a>
          <button className="sidebar-toggler" onClick={() => setCollapsed((v) => !v)}>
            <span className="material-symbols-rounded">chevron_left</span>
          </button>
        </header>

        <nav className="sidebar-nav">
          <ul className="nav-list primary-nav">
            {NAV_PRIMARY.map((item) => (
              <li key={item.label} className={`nav-item${item.children ? " dropdown-container" + (openDropdown === item.label ? " open" : "") : ""}`}>
                <a
                  href={item.children ? undefined : item.link}
                  className={`nav-link${item.children ? " dropdown-toggle" : ""}`}
                  onClick={item.children ? (e) => { e.preventDefault(); toggleDropdown(item.label); } : undefined}
                >
                  <span className="material-symbols-rounded">{item.icon}</span>
                  <span className="nav-label">{item.label}</span>
                  {item.children && (
                    <span className="dropdown-icon material-symbols-rounded">keyboard_arrow_down</span>
                  )}
                </a>
                {item.children && (
                  <ul
                    className="dropdown-menu"
                    style={{ height: openDropdown === item.label ? `${(item.children.length + 1) * 40}px` : 0 }}
                  >
                    <li className="nav-item"><a className="nav-link dropdown-title">{item.label}</a></li>
                    {item.children.map((child) => (
                      <li key={child} className="nav-item">
                        <a href={child} className="nav-link dropdown-link">{child}</a>
                      </li>
                    ))}
                  </ul>
                )}
              </li>
            ))}
          </ul>

          <ul className="nav-list secondary-nav">
            {NAV_SECONDARY.map((item) => (
              <li key={item.label} className="nav-item">
                <a href={item.link} className="nav-link">
                  <span className="material-symbols-rounded">{item.icon}</span>
                  <span className="nav-label">{item.label}</span>
                </a>
              </li>
            ))}
          </ul>
        </nav>
      </aside>
    </>
  );
}
