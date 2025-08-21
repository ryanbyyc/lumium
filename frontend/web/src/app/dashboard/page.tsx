import { auth, signOut } from "@/lib/auth"

export default async function DashboardPage() {
  const session = await auth()
  return (
    <main className="p-6 text-white">
      <h1 className="text-2xl font-semibold">Dashboard</h1>
      <p className="mt-2 text-white/80">
        Hello {session?.user?.name ?? "friend"}{" "}
        {session && `(role: ${(session as any).role ?? "n/a"})`}
      </p>
      <form
        action={async () => {
          "use server"
          await signOut({ redirectTo: "/login" })
        }}
      >
        <button className="mt-4 rounded border border-white/10 bg-white/10 px-3 py-2">
          Sign out
        </button>
      </form>
    </main>
  )
}
