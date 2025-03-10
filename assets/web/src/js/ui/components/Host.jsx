export function Host() {
  return (
    <div className="me-4 align-self-center">
      <span className="navbar-text font-monospace text-white">
        <i className="fa-solid fa-computer"></i> {varnishmon.storage.hostname}
      </span>
    </div>
  );
}
