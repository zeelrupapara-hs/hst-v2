import useLiveSymbolStore from '../../../store/useLiveSymbolStore';

const Spread = ({ symbolId }) => {
    const spread = useLiveSymbolStore((state) => state.liveSymbols?.[symbolId]?.newSpread);

    return spread
}

export default Spread