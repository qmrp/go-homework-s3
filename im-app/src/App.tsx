import { ImProvider, useIm } from "@/im/context";
import { ChatShell } from "@/app/ChatShell";
import { LoginView } from "@/app/LoginView";

function App() {
  return (
    <ImProvider>
      <AppInner />
    </ImProvider>
  );
}

export default App;

function AppInner() {
  const { state } = useIm();
  if (state.auth !== "logged_in" || !state.me) return <LoginView />;
  return <ChatShell />;
}
