import { Navbar } from "@/components/custom/Navbar/Navbar";

export default function Dashboard({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen">
      <Navbar />
      <main className="container mx-auto px-4 py-4">{children}</main>
    </div>
  );
}
