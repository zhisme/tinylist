import { Link } from 'preact-router/match';

export function Layout({ children, onLogout }) {
  return (
    <div class="min-h-screen flex">
      {/* Sidebar */}
      <aside class="w-64 bg-gray-800 text-white flex flex-col">
        <div class="p-4">
          <h1 class="text-xl font-bold">TinyList</h1>
          <p class="text-gray-400 text-sm">Email List Manager</p>
        </div>
        <nav class="mt-4 flex-1">
          <NavLink href="/">Dashboard</NavLink>
          <NavLink href="/subscribers">Subscribers</NavLink>
          <NavLink href="/campaigns">Campaigns</NavLink>
          <NavLink href="/settings">Settings</NavLink>
        </nav>
        {onLogout && (
          <div class="p-4 border-t border-gray-700">
            <button
              onClick={onLogout}
              class="w-full px-4 py-2 text-sm text-gray-300 hover:text-white hover:bg-gray-700 rounded transition-colors"
            >
              Sign out
            </button>
          </div>
        )}
      </aside>

      {/* Main content */}
      <main class="flex-1 p-8 overflow-auto">
        {children}
      </main>
    </div>
  );
}

function NavLink({ href, children }) {
  return (
    <Link
      href={href}
      class="block px-4 py-2 hover:bg-gray-700 transition-colors"
      activeClassName="bg-gray-700 border-l-4 border-blue-500"
    >
      {children}
    </Link>
  );
}
