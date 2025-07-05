interface OrcaIconProps {
  className?: string;
}

export function OrcaIcon({ className = "h-8 w-8" }: OrcaIconProps) {
  return (
    <svg 
      className={className}
      width="128" 
      height="128" 
      viewBox="0 0 128 128"
      fill="currentColor"
      xmlns="http://www.w3.org/2000/svg"
    >
      <defs>
        <clipPath id="b">
          <rect width="128" height="128"/>
        </clipPath>
      </defs>
      <g id="a" clipPath="url(#b)">
        <g transform="translate(-47.5 -5.584)">
          <path d="M899.881-2757.868c-6.845-3.385-6.653-5.126,2.445-5.375,2.7-.072,17.863.822,21.648,6.3C927.327-2752.094,906.332-2754.681,899.881-2757.868Zm53.379,9.953q-15.6-22.4-40.707-24.372c.009,0,.027-.008.037-.008-18.48-10.262-26.125-22.115-49.785-23.227q19.407,19.38,8.013,31.247-5.911,2.392-12.147,5.48-19.173,9.187-33.337,29.014c49.409-40.918,51.007-14.443,84.115-14.886,23.965-.323,39.4,2.756,41.562,1.7,1.819-.895,2.572-2.539,2.251-4.949" transform="translate(-777.832 2832.235)"/>
        </g>
      </g>
    </svg>
  );
}